package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/distribution/context"
	"github.com/go-redis/redis/v8"
	"github.com/kr/pretty"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cast"
	"github.com/urfave/cli"
	// "github.com/k0kubun/pp"
)

func main() {
	var addr string
	var db int
	var password string
	var ssl bool

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "addr, a",
			Value:       "127.0.0.1:6379",
			Usage:       "`address` for the redis-server to connect",
			Destination: &addr,
		},
		cli.IntFlag{
			Name:        "db, n",
			Value:       0,
			Usage:       "`db` for the redis-server to connect",
			Destination: &db,
		},
		cli.StringFlag{
			Name:        "password, p",
			Value:       "",
			Usage:       "`password` for the redis-server to connect",
			Destination: &password,
		},
		cli.BoolFlag{
			Name:        "ssl, s",
			Usage:       "ssl for the redis-server to connect",
			Destination: &ssl,
		},
	}

	app.Action = func(c *cli.Context) error {
		log.Printf("Connected to redis-server addr: %v, db: %v, password: %v", addr, db, password)
		opts := &redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}
		if ssl {
			opts.TLSConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		client := redis.NewClient(opts)
		defer client.Close()

		for {
			validate := func(input string) error {
				return nil
			}

			prompt := promptui.Prompt{
				Label:    fmt.Sprintf("%v[%v]>", addr, db),
				Validate: validate,
			}

			input, err := prompt.Run()

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return err
			}
			args := strings.Fields(input)
			res, err := processRedisCli(client, args...)
			if err != nil {
				fmt.Printf("cmd failed %v\n", err)
				return err
			}

			fmt.Printf("%v\n", res)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func processRedisCli(client *redis.Client, args ...string) (string, error) {
	switch args[0] {
	case "zrange":
		s, err := client.ZRange(context.Background(), args[1], cast.ToInt64(args[2]), cast.ToInt64(args[3])).Result()
		if err != nil {
			return "", err
		}
		return pretty.Sprint(s), nil
	case "zrangewithscores":
		s, err := client.ZRangeWithScores(context.Background(), args[1], cast.ToInt64(args[2]), cast.ToInt64(args[3])).Result()
		if err != nil {
			return "", err
		}
		return pretty.Sprint(s), nil
	case "zrangebyscore":
		s, err := client.ZRangeByScore(context.Background(), args[1], &redis.ZRangeBy{
			Min:    args[2],
			Max:    args[3],
			Offset: cast.ToInt64(args[4]),
			Count:  cast.ToInt64(args[5]),
		}).Result()
		if err != nil {
			return "", err
		}
		return pretty.Sprint(s), nil
	case "zrangebyscorewithscores":
		s, err := client.ZRangeByScoreWithScores(context.Background(), args[1], &redis.ZRangeBy{
			Min:    args[2],
			Max:    args[3],
			Offset: cast.ToInt64(args[4]),
			Count:  cast.ToInt64(args[5]),
		}).Result()
		if err != nil {
			return "", err
		}
		return pretty.Sprint(s), nil
	default:
		newArgs := make([]interface{}, len(args))
		for i, v := range args {
			newArgs[i] = v
		}
		s, err := client.Do(context.Background(), newArgs...).Result()
		if err != nil {
			return "", err
		}
		return pretty.Sprint(s), nil
	}
}
