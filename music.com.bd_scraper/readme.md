### Scraping songs from music.com.bd
Downloads all songs by a band/artist and saves them in home/Music directory
#### Syntax:
```
$ python scrape.py band/artist album1 album2 ...
```
If no album is specified all the albums will be downloaded. If there are spaces in the names, replace them with '_'. Names are not case sensitive.
#### Example:
```
$ python scrape.py shironamhin
$ python scrape.py arnob doob
$ python scrape.py shironamhin bondho_janala
$ python scrape.py stoic_bliss kolponar_baire light_years_ahead
$ python scrape.py AuRThoHiN AushoMAPTO_1 tRIMAtrik sUMon_O_Aurthohin Dhrubok
```