require 'open-uri'
require 'json'
require 'nokogiri'

url = "https://www.bbc.co.uk/news"

html_doc = open(url).read
doc = Nokogiri::HTML5(html_doc)
doc.errors.each do |error|
    puts error
end

top_stories = doc.xpath("//h3")

top_stories.each do |top_story|
  puts top_story.text
end