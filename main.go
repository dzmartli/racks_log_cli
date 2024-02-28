package main

import (
    "fmt"
    "strings"
    "flag"
    "log"
    "time"
    "os"
    "regexp"
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "encoding/json"
    "text/tabwriter"
    "io"
)

const (
    layoutISO string = "2006-01-02T15:04:05.000Z"
    hiRedString string = "\x1b[91m %s \x1b[0m"
    hiYellowString string = "\x1b[93m %s \x1b[0m"
    hiGreenString string = "\x1b[92m %s \x1b[0m"
)
var (
    mongoURI string = os.Getenv("MONGODB_URI")
    databaseName string = os.Getenv("MONGODB_NAME")
    collectionName string = os.Getenv("MONGODB_COLL_NAME")
)

func main() {
    defaultRange := getDefaultDateRange()
    lastFlag, rangeFlag, actionFlag := getFlags(defaultRange)
    checkRangeLenth(rangeFlag)
    checkRangeRegexp(rangeFlag)
    startDateParse, endDateParse := getDatesForFilter(rangeFlag)
    filter := getFilter(startDateParse, endDateParse)
    logs := getDataFromMongo(filter, lastFlag) 
    messages := getMessages(logs)
    printSortedMessages(messages, actionFlag)
}

// Default date range
func getDefaultDateRange() string {
    today := time.Now()
    tomorrow := time.Now().AddDate(0, 0, 1)
    defaultRange := today.Format("2006-01-02") + 
        "_" + tomorrow.Format("2006-01-02")
    return defaultRange
}

// Get falgs
func getFlags(defaultRange string) (uint, string, string) {
    lastFlag := flag.Uint(
        "last", 
        100, 
        "number of entries, has less priority than -range flag")
    rangeFlag := flag.String(
        "range", 
        defaultRange, 
        "range of dates in YYYY-MM-DD_YYYY-MM-DD format")
    actionFlag := flag.String(
        "action", 
        "all", 
        "entry action all|add|update|delete")
    flag.Parse()
    return *lastFlag, *rangeFlag, *actionFlag
}

// Check lenth of range flag
func checkRangeLenth(rangeFlag string) {
    if len([]rune(rangeFlag)) != 21 {
        fmt.Println("Wrong date range")
        os.Exit(1)
    }
}

// Range flag regex check
func checkRangeRegexp(rangeFlag string) {
    matched, _ := regexp.MatchString(
        `\d{4}-\d{2}-\d{2}_\d{4}-\d{2}-\d{2}`, 
        rangeFlag)
    if matched == false {
        fmt.Println("Wrong date range")
        os.Exit(1)
    }
}

// Dates for date range
func getDatesForFilter(rangeFlag string) (time.Time, time.Time) {
    dates := strings.Split(rangeFlag, "_")
    startDate := dates[0] + "T00:00:00.000Z"
    endDate := dates[1] + "T00:00:00.000Z"
    startDateParse, _ := time.Parse(layoutISO, startDate) 
    endDateParse, _ := time.Parse(layoutISO, endDate)
    return startDateParse, endDateParse
}

// Filter for date range
func getFilter(startDateParse time.Time, endDateParse time.Time) bson.D {
    filter := bson.D{
        {"$and",
            bson.A{
                bson.D{{"created", bson.D{{"$gt", startDateParse}}}},
                bson.D{{"created", bson.D{{"$lt", endDateParse}}}},
            },
        },
    }
    return filter
}

// Monobgodb client connection
func getDataFromMongo(filter bson.D, lastFlag uint) *[]bson.M {
    client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)
    demoDB := client.Database(databaseName)
    mongologCollection := demoDB.Collection(collectionName)
    opts := options.
                    Find().
                    SetSort(bson.D{{"created", 1}}).
                    SetLimit(int64(lastFlag))
    cursor, err := mongologCollection.Find(ctx, filter, opts)
    if err != nil {
        log.Fatal(err)
    }
    var logs []bson.M
    if err = cursor.All(ctx, &logs); err != nil {
        log.Fatal(err)
    }
    return &logs
}

// Get msg maps from log entries
func getMessages(logs *[]bson.M) *[]bson.M {
    var msg []bson.M
    for _, doc := range *logs {
        for key, value := range doc {
            if key == "msg" {
                val := value.(bson.M)
                msg = append(msg, val)
            }
        }
    }
    return &msg
}

// Sort entries by action and print them
func printSortedMessages(messages *[]bson.M, action string) {
    var (
        separator string = strings.Repeat("-", 100)
        writer io.Writer = tabwriter.NewWriter(
            os.Stdout, 10, 0, 2, ' ', tabwriter.Debug)
    )
    for _, msg := range *messages {
        if (msg["action"].(string) == "delete") && 
            (action == "all" || action == "delete") {
            printDelete(&msg, separator, writer)
        } else if (msg["action"].(string) == "add") && 
            (action == "all" || action == "add") {
            printAdd(&msg, separator, writer)
        } else if (msg["action"].(string) == "update") && 
            (action == "all" || action == "update") {
            printUpdate(&msg, separator, writer)
        } else {
            fmt.Printf("")
        }
    }
}

// Print entries with "delete" action
func printDelete(message *bson.M, separator string, writer io.Writer) {
    var (
        msg bson.M = *message
        time string = msg["time"].(string)
        user string = msg["user"].(string)
        action string = msg["action"].(string)
        modelName string = msg["model_name"].(string)
        objectName string = msg["object_name"].(string)
        pk string = msg["pk"].(string)
    )
    fmt.Fprint(writer, "\nDATE\tUSER\tACTION\tMODEL NAME\tPK\n")
    fmt.Fprint(
        writer, time, "\t", user, "\t", 
        action, "\t", modelName, "\t", pk, "\n\n")
    fmt.Fprintf(os.Stdout, hiRedString, "\tOBJECT NAME:")
    fmt.Fprint(os.Stdout, objectName, "\n\n")
    fmt.Println(separator)
}

// Print entries with "add" action
func printAdd(message *bson.M, separator string, writer io.Writer) {
    var (
        msg bson.M = *message
        time string = msg["time"].(string)
        user string = msg["user"].(string)
        action string = msg["action"].(string)
        modelName string = msg["model_name"].(string)
        fk string = msg["fk"].(string)
        newData bson.M = msg["new_data"].(bson.M)
    )
    marshalNewData, _ := json.MarshalIndent(newData, "", "    ")
    fmt.Fprint(writer, "\nDATE\tUSER\tACTION\tMODEL NAME\tFK\n")
    fmt.Fprint(
        writer, time, "\t", user, "\t", 
        action, "\t", modelName, "\t", fk, "\n\n")
    fmt.Fprintf(os.Stdout, hiGreenString, "\tNEW DATA:\n")
    fmt.Fprint(writer, string(marshalNewData), "\n\n")
    fmt.Println(separator)
}

// Print entries with "update" action
func printUpdate(message *bson.M, separator string, writer io.Writer) {
    var (
        msg bson.M = *message
        time string = msg["time"].(string)
        user string = msg["user"].(string)
        action string = msg["action"].(string)
        modelName string = msg["model_name"].(string)
        pk string = msg["pk"].(string)
        newData bson.M = msg["new_data"].(bson.M)
        oldData bson.M = msg["old_data"].(bson.M)
    )
    marshalOldData, _ := json.MarshalIndent(oldData, "", "    ")
    marshalNewData, _ := json.MarshalIndent(newData, "", "    ")
    fmt.Fprint(writer, "\nDATE\tUSER\tACTION\tMODEL NAME\tPK\n")
    fmt.Fprint(
        writer, time, "\t", user, "\t", 
        action, "\t", modelName, "\t", pk, "\n\n")
    fmt.Fprintf(os.Stdout, hiYellowString, "\tOLD DATA:\n")
    fmt.Fprint(writer, string(marshalOldData), "\n\n")
    fmt.Fprintf(os.Stdout, hiGreenString, "\tNEW DATA:\n")
    fmt.Fprint(writer, string(marshalNewData), "\n\n")
    fmt.Println(separator)
}
