#!/usr/bin/env ruby

def nth_prime(n)
  prime_numbers = []
  i = 2
  
  while prime_numbers.length < n
    if is_prime?(i)
      prime_numbers << i
    end
    i += 1
  end
  prime_numbers.last
end

def is_prime?(num)
  (2...num).each do |divisor|
    return false if num % divisor == 0
  end

  true
end

puts "Please enter a number: "
num = gets.chomp.to_i
nth_prime_num = nth_prime(num)

puts "The #{num}th prime number is #{nth_prime_num}."