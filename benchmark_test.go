package main

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

func BenchmarkBasic(b *testing.B) {
	conn, cb := setup(b, 10)
	defer cb()

	tset := map[string]string{
		"line01": "What shall I do when the Summer troubles—",
		"line02": "What, when the Rose is ripe—",
		"line03": "What when the Eggs fly off in Music",
		"line04": "From the Maple Keep?",
		"line05": "What shall I do when the Skies a'chirrup",
		"line06": "Drop a Tune on me—",
		"line07": "When the Bee hangs all Noon in the Buttercup",
		"line08": "What will become of me?",
		"line09": "Oh, when the Squirrel fills His Pockets",
		"line10": "And the Berries stare",
		"line11": "How can I bear their jocund Faces",
		"line12": "Thou from Here, so far?",
		"line13": "'Twouldn't afflict a Robin—",
		"line14": "All His Goods have Wings—",
		"line15": "I—do not fly, so wherefore",
		"line16": "My Perennial Things?",
	}
	for k, v := range tset {
		if _, err := conn.Do("SET", k, v); err != nil {
			b.Fatalf("SET err: %s", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, err := redis.String(conn.Do("GET", "line06"))
		if err != nil {
			b.Fatalf("GET err: %s", err)
		}
		if have, want := v, tset["line06"]; have != want {
			b.Fatalf("have: %s, want: %s", have, want)
		}
	}
}
