package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
)

type GeneralInfo struct {
	City        string
	WeatherInfo []Weather
}

type Weather struct {
	Day        string `json:"day"`
	Date       string `json:"date"`
	Condition  string `json:"condition"`
	Temp_day   string `json:"temp_day"`
	Temp_night string `json:"temp_night"`
}

var cityUrls = map[string]string{
	"moscow":           "https://yandex.ru/pogoda/moscow",
	"saint-petersburg": "https://yandex.ru/pogoda/saint-petersburg",
	"krasnodar":        "https://yandex.ru/pogoda/?lat=45.03547287&lon=38.97531509",
}

func main() {

	e := echo.New()
	e.GET("/:city", getWeather)

	log.Fatal(e.Start(":8080"))

}

func getWeather(c echo.Context) error {
	city := c.Param("city")

	generalProdcast := parseWeatherInfo(city, cityUrls[city])

	return c.JSON(http.StatusOK, generalProdcast)
}

func parseWeatherInfo(city, url string) GeneralInfo {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	if response.StatusCode >= 400 {
		fmt.Println("Status code: ", response.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var days []string
	var dates []string
	var temperatures [][]string
	var conditions []string

	const days_per_week = 7

	re := regexp.MustCompile(".[0-9][0-9]?")

	doc.Find("ul.swiper-wrapper").Find("li.forecast-briefly__day").Find("a.link").EachWithBreak(func(index int, item *goquery.Selection) bool {

		day_name := item.Find("div.forecast-briefly__name").Text()
		days = append(days, day_name)

		date := item.Find("time.time").Text()
		dates = append(dates, date)

		temp := item.Find("div.temp").Find("span.temp__value").Text()
		t := re.FindAllString(temp, -1)
		temperatures = append(temperatures, t)

		condition := item.Find("div.forecast-briefly__condition").Text()
		conditions = append(conditions, condition)

		if index-1 == days_per_week {
			return false
		}
		return true
	})

	fmt.Println(city)

	var temp_day []string
	var temp_night []string
	for i := 0; i < len(temperatures); i++ {
		temp_day = append(temp_day, temperatures[i][0])
		temp_night = append(temp_night, temperatures[i][1])
	}

	prodcast := []Weather{}

	for i := 0; i < len(days); i++ {
		prodcast = append(prodcast, Weather{Day: days[i], Date: dates[i], Condition: conditions[i], Temp_day: temp_day[i], Temp_night: temp_night[i]})
	}

	return GeneralInfo{City: city, WeatherInfo: prodcast}

}
