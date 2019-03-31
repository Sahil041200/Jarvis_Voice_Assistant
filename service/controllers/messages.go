package controllers

import (
	"net/http"
	"strings"
	"encoding/json"
	"fmt"
)

type response struct {
	username string
	message string
}

type statusCode struct {
	status string
}

type messageQueryBody struct {
	Head string `json:"head"`
	Link string `json:"link"`
}

type jsonResponseQuery struct {
	Status bool	`json:"status"`
	Message string `json:"message"`
	Result []messageQueryBody `json:"result"`
}

type jsonResponseWeather struct {
	Status bool	`json:"status"`
	Message string `json:"message"`
	Result string `json:"result"`
}

type weatherStr struct {
	Time string `json:"time"`
	City string `json:"city"`
	Temperature string `json:"temperature"`
	DewPoint string `json:"dew_point"`
	Humidity string `json:"humidity"`
	Visibility string `json:"visibility"`
	FeelsLike string `json:"feels_like"`
}

// MessagesController controls messages handling
func MessagesController(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	r.ParseForm()

	request := response{
		username: r.FormValue("username"),
		message: r.FormValue("message"),
	}
	fmt.Println(request)

	routes(request, w)

}

func routes(routeObject response, w http.ResponseWriter) {

	message := routeObject.message
	messageArr := strings.Split(message, " ")
	// messageTemp := message
	var firstPars string
	if strings.Contains(message, " ") {
		firstPars = message[:strings.Index(message, " ")]
	} else {
		firstPars = message
	}

	strArr := strings.Split(firstPars, " ")
	strArrDiff := strings.Split(message, " ")

	messageExceptFirstPars := strings.Join(stringDifference(strArr, strArrDiff), " ")
	// lastParsArr := strings.Split(messageTemp, " ")
	// lastPars := lastParsArr[len(lastParsArr) - 1]

	// single word operations

	if strings.ToLower(firstPars) == "google" { // for google search

		query := "https://www.google.co.in/search?q=" + messageExceptFirstPars
		result := HandlerGoogle("GET", query)

		// processing

		response := processGoogleResponses(result)
		responseJSON := jsonResponseQuery{
			Status: true,
			Message: "here are the top search results",
			Result: response,
		}
		jData, _ := json.Marshal(responseJSON)
		w.Write(jData)
		TextToSpeech(responseJSON.Message, 0)

	} else if strings.ToLower(firstPars) == "yahoo" {
		query := "https://in.search.yahoo.com/search?p=" + messageExceptFirstPars
		result := HandlerYahoo("GET", query)

		// processing

		response := processYahooResponses(result)
		responseJSON := jsonResponseQuery{
			Status: true,
			Message: "here are the top search results",
			Result: response,
		}
		jData, _ := json.Marshal(responseJSON)
		w.Write(jData)
		TextToSpeech(responseJSON.Message, 0)

	} else if strings.ToLower(firstPars) == "bing" {
		query := "https://www.bing.com/search?q=" + messageExceptFirstPars
		HandlerBing("GET", query)
	} else if strings.ToLower(firstPars) == "weather" {

		city := messageArr[len(messageArr)-2]
		state := messageArr[len(messageArr)-1]
		result := HandlerWeather(city, state)
		stringified, _ := json.Marshal(processWeather(result))
		response := jsonResponseWeather{
			Status: true,
			Message: "here are the current weather conditions",
			Result: string(stringified),
		}
		jData, _ := json.Marshal(response)
		w.Write(jData)
		TextToSpeech(response.Message + city + " " + state, 0)

	} else {
		w.Write([]byte(`{"status": "success", "message": "Hi from reply bot", "result": ""}`))
	}

}

// gives the difference of two string arrays as an array of the differed element
func stringDifference(slice1 []string, slice2 []string) []string {
    var diff []string

    // Loop two times, first to find slice1 strings not in slice2,
    // second loop to find slice2 strings not in slice1
    for i := 0; i < 2; i++ {
        for _, s1 := range slice1 {
            found := false
            for _, s2 := range slice2 {
                if s1 == s2 {
                    found = true
                    break
                }
            }
            // String not found. We add it to return slice
            if !found {
                diff = append(diff, s1)
            }
        }
        // Swap the slices, only if it was the first loop
        if i == 0 {
            slice1, slice2 = slice2, slice1
        }
    }

    return diff
}

// processes google query result, scraps the required data and returns it
func processGoogleResponses(result string) []messageQueryBody {

	subsl := "<h3 class=\"LC20lb\">"
	subsl2 := "</h3>"
	subsl3 := "<cite"
	lensubsl3 := len(subsl3)
	subsl4 := "</cite>"
	lensubsl4 := len(subsl4)
	var queryResult messageQueryBody
	var queryResultArray []messageQueryBody
	for i := 0; i < len(result) - len(subsl); i++ {
		mess := ""
		if result[i : i + len(subsl)] == subsl {
			length := i + len(subsl)
			var last int
			for j:=1; ; j++ {
				if result[length + j: length + j + len(subsl2)] == subsl2 {
					mess = result[length: length + j]
					queryResult.Head = mess
					last = length + j + len(subsl2)
					i = last
					break
				}
			}

			found := false
			for j:= 1; ; j++ {
				if result[last + j: last + j + lensubsl3] == subsl3 { // matched found for "<cite"
					for k:= 1; ; k++ {
						if result[last + j + lensubsl3 + k: last + j + lensubsl3 + k + lensubsl4] == subsl4 { // finding index for "</cite>"
							link := result[last + j + lensubsl3 + 15 : last + j + lensubsl3 + k]
							i = last + j + lensubsl3 + k + lensubsl4
							found = true
							queryResult.Link = link
							break
						}
					}
				}
				if found {
					queryResultArray = append(queryResultArray, queryResult)
					break
				}
			}
		}
	}

	return queryResultArray
}

func processWeather(response string) weatherStr  {

	fmt.Println("this is the response")
	fmt.Println(response)
	subl := "in json format"
	sublLen := len(subl)
	found := false
	var weather []byte
	var weatherInJSON weatherStr
	for i:=0; i< len(response) - sublLen; i++ {
		if response[i: i + sublLen] == subl {
			for j:=1; ; j++ {
				if response[i+sublLen+j: i+sublLen+j + 1] == "}" {
					weather = []byte(response[i+sublLen+2: i+sublLen+j+1])
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
	if !found {
		fmt.Println("corrupted logging!")
	}
	fmt.Println(string(weather))
	err := json.Unmarshal(weather, &weatherInJSON)
	if err != nil {
		panic(err)
	}
	fmt.Println(weatherInJSON)
	return weatherInJSON
}

// processes yahoo query result, scraps the required data and returns it
func processYahooResponses(result string) []messageQueryBody {

	subsl := "<a class=\" ac-algo fz-l ac-21th lh-24\"";
	subsl2 := "</a>"
	subsl3 := "<span class=\" fz-ms fw-m fc-12th wr-bw lh-17\">"
	lensubsl3 := len(subsl3)
	subsl4 := "</span>"
	lensubsl4 := len(subsl4)

	var queryResult messageQueryBody
	var queryResultArray []messageQueryBody
	for i := 0; i < len(result) - len(subsl); i++ {
		mess := ""
		if result[i : i + len(subsl)] == subsl {
			length := i + len(subsl)
			var last int
			var start int

			for k := 1; ; k++ {
				if result[length + k: length+k+1 ] == ">" {
					start =  length + k + 1;
					break;
				}
			}

			for j:=1; ; j++ {
				if result[start + j: start + j + len(subsl2)] == subsl2 {
					mess = result[start: start + j]
					queryResult.Head = mess
					last = start + j + len(subsl2)
					i = last
					break
				}
			}

			found := false
			for j:= 1; ; j++ {
				if result[last + j: last + j + lensubsl3] == subsl3 { // matched found for "<span class=\" fz-ms fw-m fc-12th wr-bw lh-17\">"
					for k:= 1; ; k++ {
						if result[last + j + lensubsl3 + k: last + j + lensubsl3 + k + lensubsl4] == subsl4 { // finding index for "</cite>"
							link := result[last + j + lensubsl3 : last + j + lensubsl3 + k]
							i = last + j + lensubsl3 + k + lensubsl4
							found = true
							flink := strings.Replace(link, "<b>", "", -1)
							finallink := strings.Replace(flink, "</b>", "", -1)
							queryResult.Link = finallink
							break
						}
					}
				}
				if found {
					queryResultArray = append(queryResultArray, queryResult)
					break
				}
			}
		}
	}
	return queryResultArray
}

