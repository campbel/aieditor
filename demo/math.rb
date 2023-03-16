#!/usr/bin/env ruby

puts "Please enter a number: "
num = gets.chomp.to_i

def prime?(num)
    (2..num/2).each do |i|
        if num % i == 0
            return false
        end
    end
    true
end

if prime?(num)
    puts "#{num} is a prime number!"
else
    puts "#{num} is not a prime number!"
end