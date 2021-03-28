package main

import (
    "net/http"
    "encoding/json"
    // "strings"
    "strconv"
    "github.com/gorilla/mux"
    "log"
    // "fmt"
    "time"
    "sync"
)

/*
section: 1
util functions
*/
func checktimeformat(time_ string)bool{
    _, err := time.Parse("15:04",time_)
    if err != nil {
        return false
    } else if time_[3]!='0'{
        return false
    } else if time_[4]!='0'{
        return false
    } else {
        return true
    }
}
func checkdateformat(date string)bool{
    _, err := time.Parse("02-01-2006", date)
    if err != nil {
        return false
    } else {
        return true
    }
}
func gettime(t int)string{
    if(t>=0&&t<=9){
        return "0"+strconv.Itoa(t)+":00"
    }
    return strconv.Itoa(t)+":00"
}
func getkey(date string,time_ string)string{
    return date + " " + time_
}
func generatekeysmonth() []string{
    keys := []string{}
    for i:=0;i<24*30;i++ {
		t := time.Now().Add(time.Duration(-i)*time.Hour)
		v := t.Format("02-01-2006 15")
		keys = append(keys,v+":00")
	}
    return keys
}
func getDate(key string) string {
	return key[:10]
}
func getTime(key string) string {
	return key[11:]
}
/* 
to order resources and access, to not get deadlock.
not required because no part of code requires morethan 1 lock at a time.
func comparekeys(key1 string, key2 string)bool,bool{
    t1,_ := time.Parse("02-01-2006 15:04",key1)
	t2,_ := time.Parse("02-01-2006 15:04",key2)
	return t1.Before(t2), t1.Equal(t2)
}
*/

/*
section: 2
data structures
*/
type meeting struct{
    Title string `json:"title"`
    Group []string `json:"group"`
    Host string `json:"host"`
}

type slot struct {
    Faculty map[string]*meeting
    Mettings map[*meeting]meeting
    lock sync.Mutex
}

var calender map[string/*date time concatination*/]slot
var globallock sync.Mutex
// func create()
// func delete()
// func read()
// func update()

/*
section : 3
routes and routehandelers
*/
func updatemeetCalender(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    name := vars["id"]
    oldtime := vars["time"]
    olddate := vars["date"]

    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()
    type dandt struct{
        Date string `json:"date"`
        Time string `json:"time"`
        Title string `json:"title"`
        Group []string `json:"group"`
        Host string `json:"host"`
    }
    var reqval dandt
    err := decoder.Decode(&reqval)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(err)
        // log.Fatal(err)
        return
    }

    newdate := reqval.Date
    newtime := reqval.Time
    newtitle := reqval.Title
    newgroup := reqval.Group
    newhost := reqval.Host
    if(checkdateformat(olddate) == false ||
        checktimeformat(oldtime) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - url format should be protocol://domain//calender/{id}/block/dd-mm-yyyy/hh:00")
        return
    }
    if(checkdateformat(newdate) == false ||
        checktimeformat(newtime) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - json format should be {'date':'dd-mm-yyyy','time':'hh:00'}")
        return
    }
    oldkey := getkey(olddate,oldtime)
    newkey := getkey(newdate,newtime)

    globallock.Lock()
    _,ok1 := calender[oldkey]
    if(ok1 == false){
        calender[oldkey] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    c,_ := calender[oldkey]
    globallock.Unlock()

    c.lock.Lock()
    v,ok2 := c.Faculty[name]
    if(ok2 == false){
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode("update failed: there is no previous record to update")
        c.lock.Unlock()
        return
    }
    // deleteprevious record
    if v == nil {
        delete(c.Faculty,name)
    } else {
        if name != v.Host {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode("401 - you are not the host of the meeting to update")
            c.lock.Unlock()
            return
        }
        for _,g := range v.Group {
            delete(c.Faculty,g)
        }
        delete(c.Mettings,v)
        delete(c.Faculty,name)
    }
    c.lock.Unlock()
    // create new record
    globallock.Lock()
    _,ok3 := calender[newkey]
    if(ok3 == false){
        calender[newkey] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    d,_ := calender[newkey]
    globallock.Unlock()
    
    d.lock.Lock()
    _,ok4 := d.Faculty[name]
    if(ok4 == true){
        w.WriteHeader(http.StatusNotAcceptable)
        json.NewEncoder(w).Encode("update failed: there is already a record at this slot")
        d.lock.Unlock()
        return
    }
    m := new(meeting)
    m.Title = newtitle
    m.Group = newgroup
    m.Host = newhost
    for _,id := range newgroup {
        _,ok5 := d.Faculty[id]
        if(ok5 == true){
            w.WriteHeader(http.StatusNotAcceptable)
            json.NewEncoder(w).Encode("error slot is already filled for "+id)
            d.lock.Unlock()
            return
        }
    }
    for _,id := range newgroup{
        d.Faculty[id] = m
    }
    d.Faculty[name] = m
    d.Mettings[m] = *m
    d.lock.Unlock()

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(reqval)
    return
}

func updateblockCalender(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    name := vars["id"]
    oldtime := vars["time"]
    olddate := vars["date"]

    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()
    type dandt struct{
        Date string `json:"date"`
        Time string `json:"time"`
    }
    var reqval dandt
    err := decoder.Decode(&reqval)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(err)
        // log.Fatal(err)
        return
    }

    newdate := reqval.Date
    newtime := reqval.Time
    if(checkdateformat(olddate) == false ||
        checktimeformat(oldtime) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - url format should be protocol://domain//calender/{id}/block/dd-mm-yyyy/hh:00")
        return
    }
    if(checkdateformat(newdate) == false ||
        checktimeformat(newtime) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - json format should be {'date':'dd-mm-yyyy','time':'hh:00'}")
        return
    }

    oldkey := getkey(olddate,oldtime)
    newkey := getkey(newdate,newtime)
    
    globallock.Lock()
    _,ok1 := calender[oldkey]
    if(ok1 == false){
        calender[oldkey] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    c,_ := calender[oldkey]
    globallock.Unlock()

    c.lock.Lock()
    v,ok2 := c.Faculty[name]
    if(ok2 == false){
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode("update failed: there is no previous record to update")
        c.lock.Unlock()
        return
    }
    // deleteprevious record
    if v == nil {
        delete(c.Faculty,name)
    } else {
        if name != v.Host {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode("401 - you are not the host of the meeting to update")
            c.lock.Unlock()
            return
        }
        for _,g := range v.Group {
            delete(c.Faculty,g)
        }
        delete(c.Mettings,v)
        delete(c.Faculty,name)
    }
    c.lock.Unlock()
    // create new record

    globallock.Lock()
    _,ok3 := calender[newkey]
    if(ok3 == false){
        calender[newkey] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    d,_ := calender[newkey]
    globallock.Unlock()

    d.lock.Lock()
    _,ok4 := d.Faculty[name]
    if(ok4 == true){
        w.WriteHeader(http.StatusNotAcceptable)
        json.NewEncoder(w).Encode("update failed: there is already a record at this slot")
        d.lock.Unlock()
        return
    }
    d.Faculty[name] = nil
    d.lock.Unlock()

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(reqval)
    return
}

func deleteCalender(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    name := vars["id"]
    date := vars["date"]
    time_ := vars["time"]

    if(checkdateformat(date) == false ||
        checktimeformat(time_) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - json format should be {'date':'dd-mm-yyyy','time':'hh:00'}")
        return
    }
    key := getkey(date,time_)

    globallock.Lock()
    _,ok1 := calender[key]
    if(ok1 == false){
        calender[key] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    c,_ := calender[key]
    globallock.Unlock()

    c.lock.Lock()
    v,ok2 := c.Faculty[name]
    if(ok2 == false){
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode("error there is no scheduled meeting there")
        c.lock.Unlock()
        return
    }
    if v == nil {
        delete(c.Faculty,name)
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode("200 - deleted. ")
        c.lock.Unlock()
        return
    }
    if name != v.Host {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode("401 - you are not the host of the meeting ")
        c.lock.Unlock()
        return
    }
    for _,g := range v.Group {
        delete(c.Faculty,g)
    }
    delete(c.Mettings,v)
    delete(c.Faculty,name)
    c.lock.Unlock()
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode("200 - deleted.. ")
    return
}

func returnMeetings(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)

    if vars["id"] != "F10" {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode("401 - you are not HOD")
        return
    }
    type meets struct {
        Date string `json:"date"`
        Time string `json:"time"`
        Meetings []meeting `json:"meetings"`
    }
    
    keys := generatekeysmonth()
    returnvalue := []meets{}
    for _,key := range keys {
        date := getDate(key)
        time_ := getTime(key)

        globallock.Lock()
        v,ok1 := calender[key]
        globallock.Unlock()

        if(ok1 == false){
            continue
        }
        appendvalue := meets{date,time_,[]meeting{}}
        v.lock.Lock()
        for _, mv := range v.Mettings {
            appendvalue.Meetings = append(appendvalue.Meetings,mv)
        }
        v.lock.Unlock()
        returnvalue = append(returnvalue,appendvalue)
    }
    json.NewEncoder(w).Encode(returnvalue)
}

func scheduleMeeting(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()
    type sc_meet struct{
        Date string `json:"date"`
        Time string `json:"time"`
        Title string `json:"title"`
        Group []string `json:"group"`
        Host string `json:"host"`
    }
    var reqval sc_meet
    err := decoder.Decode(&reqval)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(err)
        // log.Fatal(err)
        return
    }

    date := reqval.Date
    time_ := reqval.Time
    title := reqval.Title
    group := reqval.Group
    host := reqval.Host
    name := vars["id"]
    if(checkdateformat(date) == false ||
        checktimeformat(time_) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - json format should be {'date':'dd-mm-yyyy','time':'hh:00'}")
        return
    }
    key := getkey(date,time_)

    globallock.Lock()
    _,ok1 := calender[key]
    if(ok1 == false){
        calender[key] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    c,_ := calender[key]
    globallock.Unlock()
    
    c.lock.Lock()
    _,ok2 := c.Faculty[name]
    if(ok2 == true){
        json.NewEncoder(w).Encode("error slot is already filled")
        c.lock.Unlock()
        return
    }
    m := new(meeting)
    m.Title = title
    m.Group = group
    m.Host = host
    for _,id := range group {
        _,ok3 := c.Faculty[id]
        if(ok3 == true){
            json.NewEncoder(w).Encode("error slot is already filled for "+id)
            c.lock.Unlock()
            return
        }
    }
    for _,id := range group{
        c.Faculty[id] = m
    }
    c.Faculty[name] = m
    c.Mettings[m] = *m
    c.lock.Unlock()

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(reqval)
}

func blockCalender(w http.ResponseWriter, r *http.Request){
    vars := mux.Vars(r)
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()
    type dandt struct{
        Date string `json:"date"`
        Time string `json:"time"`
    }
    var reqval dandt
    err := decoder.Decode(&reqval)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(err)
        // log.Fatal(err)
        return
    }

    date := reqval.Date
    time_ := reqval.Time
    name := vars["id"]
    if(checkdateformat(date) == false ||
        checktimeformat(time_) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - json format should be {'date':'dd-mm-yyyy','time':'hh:00'}")
        return
    }
    key := getkey(date,time_)

    globallock.Lock()
    _,ok1 := calender[key]
    if(ok1 == false){
        calender[key] = slot{Faculty:make(map[string]*meeting),Mettings:make(map[*meeting]meeting)}
    }
    c,_ := calender[key]
    globallock.Unlock()

    c.lock.Lock()
    _,ok2 := c.Faculty[name]
    if(ok2 == true){
        // w.WriteHeader(http.St)
        json.NewEncoder(w).Encode("error slot is already filled")
        c.lock.Unlock()
        return
    }
    c.Faculty[name] = nil
    c.lock.Unlock()
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(reqval)
}

func returnSchedule(w http.ResponseWriter, r *http.Request){
    // return array of struct faculty
    vars := mux.Vars(r)
    type meets struct{
        Time string `json:"time"`
        Meet meeting `json:"meeting"`
    }
    type faculty struct{
        Name string `json:"name"`
        Date string `json:"date"`
        Schedule []meets `json:"schedule"`
    }
    
    if(checkdateformat(vars["date"]) == false){
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode("400 - URL should be protocol://domain/calender/dd-mm-yyyy")
        return
    }
    returnvalue := faculty{vars["id"],vars["date"],[]meets{}}
    for i:=0;i<24;i++{
        date := vars["date"]
        time_ := gettime(i)
        key := getkey(date,time_)

        globallock.Lock()
        v,ok1 := calender[key]
        globallock.Unlock()
        if(ok1 == false){
            continue
        }
        
        v.lock.Lock()
        meet, ok2 := v.Faculty[vars["id"]]
        if(ok2 == false){
            v.lock.Unlock()
            continue
        }
        if(meet == nil){
            returnvalue.Schedule = append(returnvalue.Schedule,meets{time_,meeting{"",[]string{},vars["id"]}})
            v.lock.Unlock()
            continue
        }
        meetingrecord := v.Mettings[meet]
        returnvalue.Schedule = append(returnvalue.Schedule,meets{time_,meeting{meetingrecord.Title,meetingrecord.Group,meetingrecord.Host}})
        v.lock.Unlock()
    }
    json.NewEncoder(w).Encode(returnvalue)
}

func handelrequests(){
    var router = mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/calender/{id}/block",blockCalender).Methods("POST")
    router.HandleFunc("/calender/{id}/meet",scheduleMeeting).Methods("POST")
    router.HandleFunc("/calender/{id}/{date}",returnSchedule).Methods("GET")
    router.HandleFunc("/HOD/{id}",returnMeetings).Methods("GET")
    router.HandleFunc("/calender/{id}/{date}/{time}",deleteCalender).Methods("DELETE")
    router.HandleFunc("/calender/{id}/block/{date}/{time}",updateblockCalender).Methods("PUT")
    router.HandleFunc("/calender/{id}/meet/{date}/{time}",updatemeetCalender).Methods("PUT")
    log.Fatal(http.ListenAndServe(":8070",router))
}

func main(){
    calender = make(map[string]slot)
    handelrequests()
}