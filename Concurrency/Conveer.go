package main

import (
	"fmt"
	"math/rand"
	"sync"
)


func generate(cancel <-chan struct{}) <-chan string{
	out := make(chan string)
	go func() {
		defer close(out)
		for{
			select{
			case <-cancel:
				return
			case out<-randomWord(5):
			}
		}
	}()
	return out
}

func takeUnique(cancel <-chan struct{}, in <-chan string) <-chan string {
	out := make(chan string)
	myMap := make(map[byte]bool, len(in))
	go func() {
		defer close(out)
		for{
			select{
			case <- cancel:
				return
			case word,ok := <- in:
				if !ok{
					return
				}
				mark :=true
				for i:= 0 ; i<len(word);i++{
					_,ok := myMap[word[i]]
					if !ok{
						myMap[word[i]]=true
					}else{
						mark = false
						break
					}
				}
				myMap = map[byte]bool{}
				if !mark{
					continue
				}else{
					select{
					case <-cancel:
						return
					case out<-word:
					}
			}
		}
	}
	}()
	return out
}

type Answer struct{
		mainW string
		newW string
	}

func reverse(cancel <-chan struct{}, in <-chan string) <-chan Answer{
	out := make(chan Answer)
	go func() {
		defer close(out)
		for {
			select{
			case <-cancel:
				return
			case word, ok := <-in:
				if !ok{
					return
				}
				runes := []rune(word)
				for i,j := 0, len(runes)-1;i < j; i, j = i+1, j-1{
					runes[i], runes[j] = runes[j], runes[i]
				}
				newWord:= string(runes)
				select{
				case <-cancel:
					return
				case out<-Answer{word,newWord}:
				}
			}
		
		}
		}()
	return out
}

func merge(cancel <-chan struct{}, c1, c2 <-chan Answer) <-chan Answer{
	var wg sync.WaitGroup
	wg.Add(2)
	out := make(chan Answer)
	forward := func(c <-chan Answer){
		defer wg.Done()
		for{
		select{
		case <-cancel:
			return
		case word,ok := <- c:
			if !ok{
				return
			}
			select{
			case <-cancel:
				return
			case out<-word:
				}
			}
		}
	}
	go forward(c1)
	go forward(c2)
	go func(){
		wg.Wait()
		close(out)
	}()
	return out
}

func print(cancel <-chan struct{}, in <-chan Answer, n int) {
	done := make(chan struct{})
	go func(){
		for i:=0;i<n; i++{
			select{
			case num,ok := <-in:
				if !ok{
					return
				}
				fmt.Println(num.mainW,"->", num.newW)
			case <-cancel:
				return
			}
		}
		done <- struct{}{}
	}()
	<- done
	close(done)
}

func randomWord(n int) string {
	const letters = "aeiourtnsl"
	chars := make([]byte, n)
	for i := range chars {
		chars[i] = letters[rand.Intn(len(letters))]
	}
	return string(chars)
}

func main() {
	cancel := make(chan struct{})
	defer close(cancel)

	c1 := generate(cancel)
	c2 := takeUnique(cancel, c1)
	c3_1 := reverse(cancel, c2)
	c3_2 := reverse(cancel, c2)
	c4 := merge(cancel, c3_1, c3_2)
	print(cancel, c4, 10)
}
