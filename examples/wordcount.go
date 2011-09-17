// The standard map/reduce example: counting words
// Copyright (c) 2011 Damian Gryski <damian@gryski.com>
// License: GPLv3 or, at your option, any later version
package main

import (
	"fmt"
	"strings"
	"../_obj/dmrgo"
	"strconv"
	"flag"
	"os"
)

// As example, just to show we can write our own custom protocols
type WordCountProto struct{}

func (p *WordCountProto) UnmarshalKVs(key string, values []string, k interface{}, vs interface{}) {

	kptr := k.(*string)
	*kptr = key

	vsptr := vs.(*[]int)

	v := []int{}

	for _, s := range values {
		i, _ := strconv.Atoi(s)
		v = append(v, i)
	}

	*vsptr = v
}

func (p *WordCountProto) MarshalKV(key interface{}, value interface{}) *dmrgo.KeyValue {
	ks := key.(string)
	vi := value.(int)
	return &dmrgo.KeyValue{ks, fmt.Sprintf("%d", vi)}
}

type MRWordCount struct {
	protocol dmrgo.MRProtocol // overkill -- we would normally just inline the un/marshal calls

	// mapper variables
	mappedWords int
}

func NewWordCount(proto dmrgo.MRProtocol) dmrgo.MapReduceJob {

	mr := new(MRWordCount)
	mr.protocol = proto

	return mr
}

func (mr *MRWordCount) Map(key string, value string) []*dmrgo.KeyValue {

	words := strings.Split(strings.TrimSpace(value), " ")
	kvs := make([]*dmrgo.KeyValue, len(words))
	for i, word := range words {
		mr.mappedWords++
		kvs[i] = mr.protocol.MarshalKV(word, 1)
	}

	return kvs
}

func (mr *MRWordCount) MapFinal() []*dmrgo.KeyValue {
	dmrgo.Statusln("finished -- mapped ", mr.mappedWords)
	dmrgo.IncrCounter("Program", "mapped words", mr.mappedWords)

	return []*dmrgo.KeyValue{}
}

func (mr *MRWordCount) Reduce(key string, values []string) []*dmrgo.KeyValue {

	counts := []int{}
	mr.protocol.UnmarshalKVs(key, values, &key, &counts)

	count := 0
	for _, c := range counts {
		count += c
	}

	return []*dmrgo.KeyValue{mr.protocol.MarshalKV(key, count)}
}

func main() {

	var use_proto = flag.String("proto", "json", "use protocol (json/wc)")

	flag.Parse()

	var proto dmrgo.MRProtocol

	if *use_proto == "json" {
		proto = new(dmrgo.JSONProtocol)
	} else if *use_proto == "wc" {
		proto = new(WordCountProto)
	} else {
		fmt.Println("unknown proto=", use_proto)
		os.Exit(1)
	}

	wordCounter := NewWordCount(proto)

        dmrgo.Main(wordCounter)

	os.Exit(0)
}
