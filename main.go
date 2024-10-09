package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/k0kubun/pp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"
)

func main() {
	var addr string
	var db int
	var password string
	var ssl bool

	app := cli.NewApp()

	app.Name = "redis-cli"
	app.HelpName = "redis-cli"
	app.Usage = "The portable redis-cli command lint tool"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "addr, a",
			Value:       "127.0.0.1:6379",
			Usage:       "`address` for the redis-server to connect",
			Destination: &addr,
		},
		&cli.IntFlag{
			Name:        "db, n",
			Value:       0,
			Usage:       "`db` for the redis-server to connect",
			Destination: &db,
		},
		&cli.StringFlag{
			Name:        "password, p",
			Value:       "",
			Usage:       "`password` for the redis-server to connect",
			Destination: &password,
		},
		&cli.BoolFlag{
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
				args := strings.Fields(input)
				if len(args) < 1 {
					return errors.New("redis cmd len < 1")
				}
				return nil
			}

			var input string

			huh.NewInput().Title(fmt.Sprintf("%v[%v]>", addr, db)).Validate(validate).Inline(true).Value(&input).Run()

			if input == "q" {
				break
			}

			args := strings.Fields(input)
			res, err := processRedisCli(client, args...)
			if err != nil {
				fmt.Printf("cmd failed %v\n", err)
				continue
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
		return pp.Sprint(s), nil
	case "zrangewithscores":
		s, err := client.ZRangeWithScores(context.Background(), args[1], cast.ToInt64(args[2]), cast.ToInt64(args[3])).Result()
		if err != nil {
			return "", err
		}
		return pp.Sprint(s), nil
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
		return pp.Sprint(s), nil
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
		return pp.Sprint(s), nil
	default:
		newArgs := make([]interface{}, len(args))
		for i, v := range args {
			newArgs[i] = v
		}
		s, err := client.Do(context.Background(), newArgs...).Result()
		if err != nil {
			return "", err
		}
		return pp.Sprint(s), nil
	}
}
