It's very easy to make some words **bold** and other words *italic* with Markdown. You can even [link to Google!](http://google.com)
#**Distributed Systems Assignment 2**
1. ##**setting up and execution:**
    * run
    ```
    **go build app.go**
    ```
    * if there are some packages unavailable install them by (example package *github.com/gorilla/mux*) then run *above command*
    ```
    **go get github.com/gorilla/mux**
    ```
2. ##**Decription about the application**
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
        2. 
