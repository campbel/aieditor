package main

import (
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	old = `
	│ #!/usr/bin/env ruby
	│ 
	│ puts "Please enter a number: "
	│ num = gets.chomp.to_i
	│ 
	│ def prime?(num)
	│     (2..num/2).each do |i|
	│         if num % i == 0
	│             return false
	│         end
	│     end
	│     true
	│ end
	│ 
	│ if prime?(num)
	│     puts "#{num} is a prime number!"
	│ else
	│     puts "#{num} is not a prime number!"
	│ end`

	new = `
	│ 
	│ #!/usr/bin/env ruby
	│ 
	│ puts "Please enter a number: "
	│ num = gets.chomp.to_i
	│ 
	│ def prime?(num)
	│     divisors = []
	│     (2..num/2).each do |i|
	│         if num % i == 0
	│             divisors.push(i)
	│         end
	│     end
	│     divisors
	│ end
	│ 
	│ divisors = prime?(num)
	│ 
	│ if divisors.any?
	│     puts "#{num} is not a prime number! Its divisors are #{divisors.join(', ')}"
	│ else
	│     puts "#{num} is a prime number!"
	│ end`
)

func main() {
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(old, new, false)

	fmt.Println(dmp.DiffPrettyText(diffs))
}
