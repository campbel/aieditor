import requests
from bs4 import BeautifulSoup
from urllib.parse import urljoin

# fetch the website
r = requests.get("https://www.bbc.com/")

# create a beautifulsoup object
soup = BeautifulSoup(r.text, "lxml")

# get list of top stories
headline_elems = soup.find_all("h3")

# print the headlines
for headline in headline_elems:
    print(headline.text.strip())
    linkElem = headline.find("a")
    if linkElem: # check if the linkElem exists
        link = linkElem.get("href")
        full_url = urljoin(r.url, link)
        print(full_url)
    print()