package main

import (        
		"fmt"
        "net/http"
		"encoding/json"

        "appengine"
        "appengine/datastore"
)

type SharedPreset struct {
        OwnerId string
		ShareId string
		Author string
        Preset string        
}

func init() {        
        http.HandleFunc("/afasPlugin/load", loadPreset)
        http.HandleFunc("/afasPlugin/store", storePreset)
}

func sharedPresetsKey(c appengine.Context, shareId string) *datastore.Key {        
        return datastore.NewKey(c, "SharedPresets", shareId, 0, nil)
}

func storePreset(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
		
		p := SharedPreset{
                OwnerId: r.FormValue("ownerId"),
                ShareId: r.FormValue("shareId"),
                Author:    r.FormValue("author"),
                Preset:    r.FormValue("preset"),
        }

        key := datastore.NewIncompleteKey(c, "Preset", sharedPresetsKey(c, p.ShareId))
        _, err := datastore.Put(c, key, &p)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
		
		writeJson(w, r, nil)
}

func loadPreset(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
		
		q := datastore.NewQuery("Preset").Ancestor(sharedPresetsKey(c, r.FormValue("shareId")))
		
		if count, err := q.Count(c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
            return
		} else {
			presets := make([]SharedPreset, 0, count)
			if _, err := q.GetAll(c, &presets); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			writeJson(w, r, presets)			
			return
		}
		
		writeJson(w, r, nil)		
		return
}

func writeJson(w http.ResponseWriter, r *http.Request, presets []SharedPreset) {
	cb := r.FormValue("callback")
	
	if presets == nil {
		if cb != "" {
			fmt.Fprintf(w, cb + "('')")			
		}
		return
	}
	
	if cb != "" {
		fmt.Fprintf(w, cb + "(")
		json.NewEncoder(w).Encode(presets)
		fmt.Fprintf(w, ");")
	} else {
		json.NewEncoder(w).Encode(presets)
	}
}