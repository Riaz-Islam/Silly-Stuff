from bs4 import BeautifulSoup, SoupStrainer
import requests
import urllib
import sys
import os
import os.path
import errno

bandname = ""
album = ""
desired = "download.music.com.bd/Music"
homedir = os.path.expanduser("~") + "/Music"
vis = dict()

class bcolors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:  # Python >2.5
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def getAllLinks(url):
    list = []
    try:
        page = requests.get(url)    
        data = page.text
        soup = BeautifulSoup(data, "html.parser")
        for link in soup.find_all('a'):
            list.append(link.get('href'))
    except:
        pass
    return list

def isinvalid(cur, depth):
    if ("music.com.bd" not in cur) or (bandname not in cur) or ("search" in cur) or ("mailto" in cur) or ("(Full Album).zip") in cur:
        return True
    if depth > 1 and len(album) > 0 and album not in cur:
        return True
    return False

def makelookgood(cur):
    if cur.startswith("http:"):
        cur = cur.replace("http:", "https:")
    if cur.endswith("/") == False:
        cur = cur + "/"
    if cur.startswith("http") == False:
        cur = "https:" + cur
    return cur

def processLink(url):
    if desired not in url:
        return
    sep = url.split("/")
    targetdir = homedir + "/" + sep[-4] + "/" + sep[-3]
    mkdir_p(targetdir)
    targetfile = targetdir + "/" + sep[-2]
    print bcolors.OKBLUE + "Downloading @" + targetfile + bcolors.ENDC
    url = url.replace("https:", "http:")
    url = url[:-1]
    try:
        urllib.urlretrieve(url, targetfile)
    except:
        print bcolors.WARNING + "Download failed" + bcolors.ENDC

def dfs(cur, depth):
    cur = makelookgood(cur)
    if isinvalid(urllib.unquote(cur), depth):
        #print "skip: ", cur
        return
    processLink(cur)
    list = getAllLinks(cur)
    vis[cur] = 1
    
    for next in list:
        nexturl = makelookgood(str(next))
        if nexturl in vis:
            continue
        dfs(nexturl, depth+1)

def getname(name):
    if len(name) == 0:
        return ""
    list = name.split("_")
    name = ""
    for word in list:
        name = name + word.capitalize() + " "
    return name[:-1]

def go():
    url = "https://www.music.com.bd/download/browse/A/Aurthohin/"
    urlband = str(bandname[0]) + "/" + bandname
    url = url.replace("A/Aurthohin", urlband)
    print "starting from: ", url
    dfs(url, 0)

if len(sys.argv) < 2:
    print bcolors.WARNING + "Please specify band name" + bcolors.ENDC
    exit(0)

bandname = getname(sys.argv[1])

if len(sys.argv) == 2:
    print bcolors.OKBLUE + "Downloading all albums of " + bandname + bcolors.ENDC
    go()
else:
    for i in range(2, len(sys.argv)):
        vis = dict()
        album = getname(sys.argv[i])
        print bcolors.OKBLUE + "Downloading album " + album + " of " + bandname + bcolors.ENDC 
        go()
