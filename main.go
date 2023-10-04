// weather_service is a demonstration project
// it requires an OpenWeather API key from https://openweathermap.org/ stored in env $OPENWEATHER_API_KEY
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

var apikey string

func init() {
	apikey = os.Getenv("OPENWEATHER_API_KEY")
	if apikey == "" {
		log.Fatal("no api key present in environment variable: 'OPENWEATHER_API_KEY'")
	}
}

const baseurl = "https://api.openweathermap.org/data/2.5/weather?units=imperial&appid="

func main() {
	var port int
	flag.IntVar(&port, "p", 3040, "Provide a port number")
	flag.Parse()

	http.HandleFunc("/", mainPage)
	http.HandleFunc("/weather", currentweather)

	fmt.Printf("Go to localhost:%v to request current weather\n", port)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, mainpage)
}

func currentweather(w http.ResponseWriter, r *http.Request) {
	lat := r.FormValue("lat")
	lon := r.FormValue("lon")

	current, err := fetchCurrentWeatherForLatLon(lat, lon)
	if err != nil {
		fmt.Fprintf(w, "unable to get weather:%v", err)
		return
	}

	tmpl, err := template.New("weathertemplate").Parse(resultpage)
	if err != nil {
		fmt.Fprintf(w, "unable to parse template:%v", err)
		return
	}
	err = tmpl.Execute(w, current)
	if err != nil {
		fmt.Fprintf(w, "unable to execute template:%v", err)
		return
	}
}

func fetchCurrentWeatherForLatLon(lat, lon string) (*SimpleWeather, error) {
	resp, err := http.Get(baseurl + apikey + "&lat=" + lat + "&lon=" + lon)
	if err != nil {
		return nil, fmt.Errorf("could not get weather: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(string(body))
	}

	current := &CurrentWeather{}
	err = json.Unmarshal(body, current)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	description := "Moderate"
	switch {
	case current.Main.FeelsLike < 40:
		description = "Cold"
	case current.Main.FeelsLike > 80:
		description = "Hot"
	}

	currentdescription := ""
	for i := range current.Weather {
		currentdescription += fmt.Sprintf("%v: %v\n", current.Weather[i].Main, current.Weather[i].Description)
	}

	name := current.Name
	if name == "" {
		name = "an unknown location"
	}

	return &SimpleWeather{
		Name:               name,
		Temp:               current.Main.FeelsLike,
		TempDescription:    description,
		CurrentDescription: currentdescription,
	}, nil
}

type SimpleWeather struct {
	Name               string
	TempDescription    string
	Temp               float64
	CurrentDescription string
}
type CurrentWeather struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		SeaLevel  int     `json:"sea_level"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

var mainpage = `<!DOCTYPE html>
<html>
<head>
<title>Weather Service</title>
</head>
<body>

<h1>Select latitude and longitude to access realtime weather at any location</h1>
<br>
<form action=/weather>
  <label for="fname">Lat:</label><br>
  <input type="text" id="lat" name="lat"><br>
  <label for="lname">Lon:</label><br>
  <input type="text" id="lon" name="lon">
  <br>
  <input type="submit" value="Submit">
</form>

</body>
</html>`

var resultpage = `<!DOCTYPE html>
<html>
<head>
<title>Weather Service</title>
</head>
<body>

<h1>Current Weather near {{.Name}}:</h1>
<h2>{{.CurrentDescription}}</h2>
<h2>{{.TempDescription}}: {{.Temp}}&deg;F</h2>

<br>
Weather data provided by <a href="https://openweathermap.org/">OpenWeather</a>

</body>
</html>`
