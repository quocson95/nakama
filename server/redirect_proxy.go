package server

import (
	"fmt"
	"io"
	"net/http"
)

func handlerRedirect(w http.ResponseWriter, r *http.Request, newhost string) {

	newReq, _ := http.NewRequest(r.Method, "http://"+newhost+"/"+r.URL.Path, r.Body)
	// newReq.Host = newhost
	// newReq.URL.Host = newhost
	// newReq.URL.Scheme = r.URL.Scheme
	query := r.URL.Query().Encode()
	newReq.URL.RawQuery = query
	for k, v := range r.Header {
		for _, r := range v {
			fmt.Printf("add header k:v %s:%s", k, r)
			newReq.Header.Add(k, r)
		}
	}
	resp, err := http.DefaultClient.Do(newReq)
	fmt.Println("do req ")
	if err != nil {
		fmt.Printf("req error %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		data, _ := io.ReadAll(resp.Body)
		fmt.Printf("body %s", string(data))
		w.Write(data)
		resp.Body.Close()
	}
	return
}
