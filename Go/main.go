package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strconv"
    "sync"

)

type Furniture struct {
    ID           int64   `json:"id"`
    Name         string `json:"name"`
    Manufacturer string `json:"manufacturer"`
    Height       float64 `json:"height"`
    Width        float64 `json:"width"`
    Length       float64 `json:"length"`
}

var furnitureData []Furniture
var nextID int64
var mutex sync.Mutex

func loadFurnitureData(filename string) {
    data, err := ioutil.ReadFile(filename) 
    if err != nil {
        if os.IsNotExist(err) {
            furnitureData = []Furniture{}
            nextID = 1
            return
        }
        log.Fatal(err)
    }
    if err := json.Unmarshal(data, &furnitureData); err != nil {
        log.Fatal(err)
    }
    if len(furnitureData) > 0 {
        lastID := furnitureData[len(furnitureData)-1].ID
        nextID = lastID + 1
    } else {
        nextID = 1
    }
}

func saveFurnitureData(filename string) {
    mutex.Lock()
    defer mutex.Unlock()
    data, err := json.MarshalIndent(furnitureData, "", "    ")
    if err != nil {
        log.Fatal(err)
    }
    if err := ioutil.WriteFile(filename, data, 0644); err != nil { 
        log.Fatal(err)
    }
}

func getFurniture(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()
    json.NewEncoder(w).Encode(furnitureData)
}

func createFurniture(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    var newFurniture Furniture
    if err := json.NewDecoder(r.Body).Decode(&newFurniture); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    newFurniture.ID = nextID
    nextID++
    furnitureData = append(furnitureData, newFurniture)
    saveFurnitureData("furniture.json")

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newFurniture)
}

func getOneFurniture(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    for _, f := range furnitureData {
        if f.ID == id {
            json.NewEncoder(w).Encode(f)
            return
        }
    }

    http.Error(w, "Furniture not found", http.StatusNotFound)
}


func updateFurniture(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var updatedFurniture Furniture
    if err := json.NewDecoder(r.Body).Decode(&updatedFurniture); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    updatedFurniture.ID = id 


    for i, f := range furnitureData {
        if f.ID == id {
            furnitureData[i] = updatedFurniture
            saveFurnitureData("furniture.json")
            json.NewEncoder(w).Encode(updatedFurniture)
            return
        }
    }

    http.Error(w, "Furniture not found", http.StatusNotFound)
}


func patchFurniture(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()
    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }


    for i, f := range furnitureData {
        if f.ID == id {
             
            patch := make(map[string]interface{})
            if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            updatedFurniture := f 

            
            for key, val := range patch {
                switch key {

                case "name":
                    if name, ok := val.(string); ok {
                        updatedFurniture.Name = name
                    }
                case "manufacturer":
                    if manufacturer, ok := val.(string); ok {
                        updatedFurniture.Manufacturer = manufacturer
                    }
                 case "height":
                    if height, ok := val.(float64); ok {
                        updatedFurniture.Height = height
                    }
                case "width":
                    if width, ok := val.(float64); ok {
                        updatedFurniture.Width = width
                    }
                 case "length":
                    if length, ok := val.(float64); ok {
                        updatedFurniture.Length = length
                    }



                }

            }
            furnitureData[i] = updatedFurniture
            saveFurnitureData("furniture.json")


            w.WriteHeader(http.StatusNoContent) 
            return
        }
    }
    http.Error(w, "Furniture not found", http.StatusNotFound)

}




func deleteFurniture(w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    defer mutex.Unlock()

    vars := mux.Vars(r)
    idStr := vars["id"]
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    for i, f := range furnitureData {
        if f.ID == id {
            furnitureData = append(furnitureData[:i], furnitureData[i+1:]...)
            saveFurnitureData("furniture.json")
            w.WriteHeader(http.StatusNoContent) 
            return
        }
    }

    http.Error(w, "Furniture not found", http.StatusNotFound)
}


func main() {
    loadFurnitureData("furniture.json")

    r := mux.NewRouter()
    r.HandleFunc("/furniture", getFurniture).Methods("GET")
    r.HandleFunc("/furniture", createFurniture).Methods("POST")
    r.HandleFunc("/furniture/{id}", getOneFurniture).Methods("GET")
    r.HandleFunc("/furniture/{id}", updateFurniture).Methods("PUT")
    r.HandleFunc("/furniture/{id}", patchFurniture).Methods("PATCH")
    r.HandleFunc("/furniture/{id}", deleteFurniture).Methods("DELETE")


    fmt.Println("Сервер запущен на порту 8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}