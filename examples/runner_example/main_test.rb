begin
  require_relative 'main'
rescue LoadError
  puts 'File main.rb not found'
  exit 1
end

expected = 'Hello, World!'

exit if hello == expected

puts "Expected #{expected.inspect}, got #{hello.inspect}"
exit 1
