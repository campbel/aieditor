puts "Please enter a number:"
num = gets.chomp.to_i

def is_prime?(num)
  for divisor in 2..(num - 1)
    if num % divisor == 0
      return false
    end
  end
  true
end

if is_prime?(num)
  puts "#{num} is prime."
else
  puts "#{num} is not prime."
end