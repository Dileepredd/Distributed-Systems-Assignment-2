**Distributed Systems Assignment 2**
1. **setting up and execution:**
    * run
    ```
    go build app.go
    ```
    * if there are some packages unavailable install them by (example package **github.com/gorilla/mux**) then run **above command**
    ```
    go get github.com/gorilla/mux
    ```
2. **Decription about the application interface**
    * URLS that are present
        ```
        Note: 
        1. here {id} is a placeholder for name of the faculty.
            eg: F1, F2, ... ,F10
        2. here {date} is a placeholder for date in dd-mm-yyyy representation.
        3. here {time} is a placeholder for time in hh:00 representation.
        ```
        1. **HTTP POST /calender/{id}/block**
            * also need to send json object of type
                ```
                {
                    "date" : "dd-mm-yyyy",
                    "time" : "hh:00"
                }
                ```
            * returns json object if operation sussess.
            * **used to blockCalender for {id} faculty**
        2. **HTTP POST /calender/{id}/meet**
            * also need to send json object of type
                ```
                {
                    "date" : "dd-mm-yyyy",
                    "time" : "hh:00",
                    "title" : "class commity meeting",
                    "group" : ["F2","F3"],
                    "host" : "F1"
                }
                ```
            * returns json object if the operation is success.
            * **used to scheduleMeeting for {id} faculty**
        3. **HTTP GET /calender/{id}/{date}**
            * returns json object of type
                ```
                {
                    "name" : "F1",
                    "date" : "dd-mm-yyyy",
                    "schedule" : [{
                        "time" : "hh:00",
                        "meet" : {
                            "title" : "class commity meeting",
                            "group" : ["F2","F3]
                            "host" : "F1"
                        }
                    }]
                }
                ```
            * **returns all sheduled meetings and blockCalenders(identified by group size == 0) on the given date of given id faculty**
        4. **HTTP GET /HOD/{id}**
            * returns json object of type
                ```
                [
                    {
                        "date" : "dd-mm-yyyy",
                        "time: : "hh:00",
                        "meetings" : [
                            {
                                "title" : "class commity meeting",
                                "group" : ["F2","F3]
                                "host" : "F1"
                            }
                        ]
                    }
                ]
                ```
            * **returns all the meetings from past one month**
        5. **HTTP DELETE /calender/{id}/{date}/{time}**
            * **used to delete blockCalender and scheduleMeeting of {id} on {date} and {time}**
        6. **HTTP PUT /calender/{id}/block/{date}/{time}**
            * **used to delete scheduleMeeting or blockCalender of {id} on {date} and {time}, and create new blockCalender**
        7. **HTTP PUT /calender/{id}/meet/{date}/{time}**
            * **used to delete scheduleMeeting or blockCalender of {id} on {date} and {time}, and create new scheduleMeeting**
3. **code explaination:**
    * section 1: contains util functions
        ```
        func checktimeformat(time_ string)bool
        // to check wether the given string is in expected time format
        ```
        ```
        func checkdateformat(date string)bool
        // to check wether the given string is in expected date format
        ```
        ```
        func gettime(t int)string
        // return string with expected time format, given time in hours as integer. 
        ```
        ```
        func getkey(date string,time_ string)string
        // return key using date and time.
        ```
        ```
        func generatekeysmonth() []string
        // return list of keys for last month
        ```
        ```
        func getDate(key string) string
        // return date from the key string.
        ```
        ```
        func getTime(key string) string
        // return time from the key string.
        ```
        ```
        // not needed, because there was no need to simultaneously access two or more resources.
        func comparekeys(key1 string, key2 string)bool,bool
        // used to order keys. so that lock aquiring is done decreasing order of the keys.
        // this will ensure that, there won't be deadlock.
        ```
    * section 2: Data structurs
        * is used to store meeting details
            ```
            type meeting struct{
                Title string `json:"title"`
                Group []string `json:"group"`
                Host string `json:"host"`
            }
            ```
        * here key for the map is concatenation of date and time
            ```
            var calender map[string/*date time concatination*/]slot
            ```
        * and value is structure
            ```
            type slot struct {
                Faculty map[string]*meeting
                Mettings map[*meeting]meeting
                lock sync.Mutex
            }
            // Faculty key is name of the faculty and value is nil or pointer.if nil it is a blockcalender else it is a schedulemeeting.
            // Meetings key is address and value is object. object contains meeting details.
            // lock is used before accessing this record, so that it avoids race conditions.
            ```
        * a global variable
            ```
            var calender map[string/*date time concatination*/]slot
            // this is the datastructure
            var globallock sync.Mutex
            // this is used before access calender variable to search and retriew record address. 
            ```
    * section 3: routes and route handlers.
        * generally route handelers contain 
            * code to convert json(received) to local type.
            * some critical code section (shown below).
            * code to generate key, validate key.
            * code to convert local type to json(to send). 
4. **critical code sections where race conditions could happen:**
    ```
    globallock.Lock()
    _,ok1 := calender[key]
    if(ok1 == false){
        calender[key] = slot{Faculty:make(map[string]*meeting)  ,Mettings:make(map[*meeting]meeting)}
    }
    c,_ := calender[key]
    globallock.Unlock()
    // where ever this code appears, it is sourounded by global lock to metigate race condition.
    // executed as a logical unit in which,
    // we search whether key is present in the map.
    // if not present insert (key,value) pair in the map.
    // retrive reference to the record.
    ```
    ```
    c.lock.Lock()
    // code to manipulate c record.
    c.lock.Unlock()
    // used to manipulate c record for inserting blockCalender or schedulemeeting etc.
    ```
5. **race conditions are mitigated by above critical section codes.**
6. **parallelism is achived by granularizing lock application on records instead on whole datastructure at the expence on memory.**