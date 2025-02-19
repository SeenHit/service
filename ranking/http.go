package main

import (
	"fmt"
	"github.com/hacash/core/fields"
	"log"
	"net/http"
	"strconv"
)

func (api *Ranking) startHttpApiService() {
	port := api.http_api_listen_port
	if port == 0 {
		// Do not start the server
		fmt.Println("config http_api_listen_port==0 do not start http api service.")
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ResponseData(w, ResponseCreateData("service", "hacash ranking service"))
	})

	// route
	mux.HandleFunc("/query", api.apiHandleFunc)

	// Set listening port
	portstr := strconv.Itoa(port)
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Hacash Ranking Service] Http api listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}

func (api *Ranking) apiHandleFunc(w http.ResponseWriter, r *http.Request) {
	api.dataChangeLocker.Lock()
	defer api.dataChangeLocker.Unlock()

	action := CheckParamString(r, "action", "")

	if action == "ranking" {

		kind := CheckParamString(r, "kind", "")
		if kind == "hacash" {

			list := make([]string, len(api.hacash_balance_ranking_100))
			for i, v := range api.hacash_balance_ranking_100 {
				percent := 0.0
				if api.current_circulation > 0 {
					percent = v.GetBalance() / api.current_circulation * 100
				}
				list[i] = fmt.Sprintf("%s %.4f %.2f", v.Address.ToReadable(), v.GetBalance(), percent)
			}
			ResponseList(w, list)

		} else if kind == "diamond" {

			list := make([]string, len(api.diamond_balance_ranking_100))
			for i, v := range api.diamond_balance_ranking_100 {
				percent := 0.0
				if api.minted_diamond > 0 {
					percent = v.GetBalance() / float64(api.minted_diamond) * 100
				}
				list[i] = fmt.Sprintf("%s %d %.2f", v.Address.ToReadable(), v.GetBalanceForceUint64(), percent)
			}
			ResponseList(w, list)

		} else if kind == "bitcoin" {

			list := make([]string, len(api.satoshi_balance_ranking_100))
			for i, v := range api.satoshi_balance_ranking_100 {
				percent := 0.0
				if api.minted_diamond > 0 {
					percent = v.GetBalance() / float64(api.transferred_bitcoin*100000000) * 100
				}
				list[i] = fmt.Sprintf("%s %.4f %.2f", v.Address.ToReadable(), v.GetBalance()/100000000, percent)
			}
			ResponseList(w, list)

		} else {
			ResponseError(w, fmt.Errorf("cannot find kind <%s>", kind))
		}
	} else if action == "account_diamonds" {

		addrstr := CheckParamString(r, "address", "")
		_, e1 := fields.CheckReadableAddress(addrstr)
		if e1 != nil {
			ResponseErrorString(w, "address format error")
			return
		}

		// Query the diamond list of the account
		diatable, hav1 := api.cache_update_diamonds[addrstr]
		if !hav1 {
			// Load from disk
			v, e := api.ldb.Get([]byte("ds"+addrstr), nil)
			//fmt.Println(string(v), e)
			if e == nil {
				diatable = v // load ok
			}
		}

		// List all diamonds
		data := ResponseCreateData("diamonds", string(diatable))
		ResponseData(w, data) // ok
	} else {
		ResponseError(w, fmt.Errorf("cannot find action <%s>", action))
	}
}
