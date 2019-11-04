import requests
import json
from bs4 import BeautifulSoup

LIST_ADDR = "https://www.vocabulary.com/lists/666067"
VOC_ADDR = "https://www.vocabulary.com/dictionary/" #vocabulary.com
WORDS_API_ADDR = "https://wordsapiv1.p.rapidapi.com/words/"
TITLE_IDENTIFIER = u"##TITLE##"
BODY_IDENTIFIER = u"##HTML_BODY##"
DUV_IDENTIFIER = u"##DIV##"
KEYA = "1c5ca6b81cmsh3"
KEYB = "d9721e16177be8p1b3db"
KEYC = "djsn3f41e351c042"

def uni(str):
    return unicode(str, "utf-8")

def get_def_from_voc(word):
    page = requests.get(VOC_ADDR + word)
    soup = BeautifulSoup(page.content, 'html.parser')
    ret = ""
    for data in soup.findAll('p',{'class':'short'}):
        ret += uni(str(data))
    for data in soup.findAll('p',{'class':'long'}):
        ret += uni(str(data))
    return ret

def get_def_from_oxford(word):
    url = WORDS_API_ADDR + word
    r = requests.get(url, headers={"x-rapidapi-host": "wordsapiv1.p.rapidapi.com", "x-rapidapi-key": KEYA+KEYB+KEYC})
    content = r.json()
    return content

def get_word_list(addr):
    page = requests.get(addr)
    soup = BeautifulSoup(page.content, 'html.parser')
    list = []
    for data in soup.findAll('a',{'class':'word dynamictext'}):
        list.append(data.text.strip())
    return list

def generate_body(word):
    dic = get_def_from_oxford(word)
    body = u"<b>" + dic['pronunciation']['all'] + u"</b><br>"
    for defn in dic['results']:
        body += u"<p><i>" + defn['partOfSpeech'] + u".</i> " + defn['definition'] + u" "
        if 'examples' in defn:
            body += '/'.join(defn['examples'])
        body += "</p>"
    body += u"<br>" + get_def_from_voc(word)
    return body

script = '''
<script type="text/javascript" 
src="https://ajax.googleapis.com/ajax/libs/jquery/1.4.4/jquery.min
.js"></script>
<script type="text/javascript">
function toggleDiv(divId) {
   $("#"+divId).toggle();
   $('.toggle').not($("#"+divId)).hide();
}
</script>
'''
template = '''
<a href="javascript:toggleDiv('##DIV##');" style="background-color: #fff;">##TITLE##</a>
<div class="toggle" id="##DIV##" style="background-color: #fff; display:none;">
##HTML_BODY##
</div>
<br /><br />
'''


print("doing ...")
html = open("index.html", "w")
html.write (script)
words = get_word_list(LIST_ADDR)
cnt = 0
for word in words:
    body = generate_body(word)
    #print(body)
    block = template.replace(DUV_IDENTIFIER, word)
    block = block.replace(TITLE_IDENTIFIER, word)
    block = block.replace(BODY_IDENTIFIER, body)
    html.write(block.encode('utf8'))
    cnt += 1
    print("done", cnt)
    if cnt == 5:
        break

print("done all")