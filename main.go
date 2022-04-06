package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/machinebox/graphql"
)

type UniswapResponse struct {
	PoolDayDatas []PoolDayData
}

type PoolDayData struct {
	FeesUSD string `json:"feesUSD"`
	Id      string `json:"id"`
	Pool    struct {
		Id string `json:"id"`
	} `json:"pool"`
	TvlUSD string `json:"tvlUSD"`
}

const startDay = "2022-01-01T00:00:00Z"
const endDay = "2022-02-28T00:00:00Z"

const graphEndpoint = "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-alt"

func main() {
	//Start a timer!
	start := time.Now()

	status := "."

	fmt.Println("Looking for Moons...")

	client := graphql.NewClient(graphEndpoint)

	var result UniswapResponse
	var lastID string
	var veryMoon string

	rates := make(map[string]float64)

	for {
		fmt.Println(status)
		err := client.Run(
			context.Background(),
			buildRequest(lastID, convertTime(startDay), convertTime(endDay)),
			&result)
		if err != nil {
			panic(err)
		}
		if len(result.PoolDayDatas) == 0 {
			break
		}
		status = status + "."
		for _, data := range result.PoolDayDatas {
			//Handle pagination
			lastID = data.Id

			//Because we can't unmarshal integers, I need to convert strings to float64
			fees := parseFloat(data.FeesUSD)
			tvl := parseFloat(data.TvlUSD)

			//Only include active pools
			if fees == 0 || tvl == 0 {
				continue
			}

			//Calculate
			rate := fees / tvl

			//Add rates to pool.id --> rate map
			if _, ok := rates[data.Pool.Id]; !ok {
				rates[data.Pool.Id] = rate
			} else {
				rates[data.Pool.Id] += rate
			}

			//Start looking for the moon
			if veryMoon == "" {
				veryMoon = data.Pool.Id
			}

			if rates[data.Pool.Id] > rates[veryMoon] {
				veryMoon = data.Pool.Id
			}
		}
	}

	//Report execution time
	duration := time.Since(start)
	fmt.Println(duration)
	fmt.Println(veryMoon)
	fmt.Println(rates[veryMoon])

}

func buildRequest(lastID string, sDate int64, eDate int64) *graphql.Request {
	query := fmt.Sprintf(`
	{
		poolDayDatas(first: 1000, where: {id_gt:"%s", date_gte:%v, date_lte:%v}) {
		  id
		  pool {
			id
		  }
		  tvlUSD
		  feesUSD
		}
	}`, lastID, sDate, eDate)
	return graphql.NewRequest(query)
}

func convertTime(datetime string) int64 {
	timestamp, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		panic(err)
	}
	return timestamp.Unix()
}

func parseFloat(jsonString string) float64 {
	float, err := strconv.ParseFloat(jsonString, 64)
	if err != nil {
		panic(err)
	}
	return float
}
