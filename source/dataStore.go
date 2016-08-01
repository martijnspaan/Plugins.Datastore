package main

import (		
		"fmt"
		"net/http"
		"encoding/json"
		//"strconv"

		"appengine"
		"appengine/datastore"
)

type SharedPreset struct {
		OwnerId string
		ShareId string
		Author string
		PresetName string
		Preset string `datastore:",noindex"`
}

func init() {		
		http.HandleFunc("/afasPlugin/load", loadPreset)
		http.HandleFunc("/afasPlugin/store", storePreset)
		http.HandleFunc("/afasPlugin/remove", removePreset)
		http.HandleFunc("/afasPlugin/getShareIds", getShareIds)
}

func sharedPresetsKey(c appengine.Context, shareId string) *datastore.Key {		
		return datastore.NewKey(c, "SharedPresets", shareId, 0, nil)
}

func storePreset(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		
		p := SharedPreset{
				OwnerId: r.FormValue("ownerId"),
				ShareId: r.FormValue("shareId"),
				Author:	r.FormValue("author"),
				PresetName:	r.FormValue("presetName"),
				Preset:	r.FormValue("preset"),
		}

		if checkIsNewOrOwner(r, p.PresetName, p.ShareId, p.OwnerId) {
			key := datastore.NewKey(c, "Preset", p.PresetName, 0, sharedPresetsKey(c, p.ShareId))
			_, err := datastore.Put(c, key, &p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		
		writeJson(w, r, nil)
}

func removePreset(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		
		ownerId := r.FormValue("ownerId")
		shareId := r.FormValue("shareId")
		presetName := r.FormValue("presetName")		

		var keys []*datastore.Key
		var err error
		
		q := datastore.NewQuery("Preset").KeysOnly().Ancestor(sharedPresetsKey(c, shareId))
		if presetName == "" {
			keys, err = q.Filter("OwnerId=", ownerId).GetAll(c, nil)
		} else {
			keys, err = q.Filter("OwnerId=", ownerId).Filter("PresetName=", presetName).GetAll(c, nil)
		}
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		err = datastore.DeleteMulti(c, keys)
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

func getShareIds(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
						
		q := datastore.NewQuery("Preset").Project("ShareId").Distinct()
		
		if count, err := q.Count(c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			shareIds := make([]string, count, count)
			
			t := q.Run(c)
			
			var i = 0
			for {				
				var preset SharedPreset
				_, err := t.Next(&preset)				
				
				if err == datastore.Done {
					break
				}
				
				shareIds[i] = preset.ShareId
				i++
			}
			
			writeJson(w, r, shareIds)			
			return
		}
		
		writeJson(w, r, nil)
}

func checkIsNewOrOwner(r *http.Request, presetName, shareId, ownerId string) bool {	
	c := appengine.NewContext(r)
	
	q := datastore.NewQuery("Preset").KeysOnly().Ancestor(sharedPresetsKey(c, shareId))	
	
	count, err := q.Filter("PresetName=", presetName).Count(c);	
	if  err == nil && count == 0 {
		// No preset with this name, so is new
		return true
	}
	
	count, err = q.Filter("OwnerId=", ownerId).Filter("PresetName=", presetName).Count(c);
	if  err == nil && count > 0 {
		// Preset exist and owned
		return true
	}
		
	return false
}

func writeJson(w http.ResponseWriter, r *http.Request, value interface{}) {

  w.Header().Set("Access-Control-Allow-Origin", "https://57439.afasinsite.nl")
  
	cb := r.FormValue("callback")
	
	if &value == nil {
		if cb != "" {
			fmt.Fprintf(w, cb + "('')")
		}
		return
	}
	
	if cb != "" {
		fmt.Fprintf(w, cb + "(")
		json.NewEncoder(w).Encode(value)
		fmt.Fprintf(w, ");")
	} else {
		json.NewEncoder(w).Encode(value)
	}
}