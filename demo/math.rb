#!/usr/bin/env ruby

def is_prime?(num)
    divisors = []
    (2...num).each do |divisor|
        divisors << divisor if num % divisor == 0
    end 
    divisors
end

puts "Please enter a number: "
num = gets.chomp.to_i

if is_prime?(num).empty?
    puts "#{num} is a prime number!"
else
    puts "#{num} is not a prime number! Its divisors are: #{is_prime?(num).join(", ")}"
end