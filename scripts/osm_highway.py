import csv
import time
import requests


def getSegmentClassification(ways):
    try:
        if 0 == len(ways):
            return []
        results = []
        url = 'http://overpass-api.de/api/interpreter?data=[out:json];way(id:{0});out tags;'.format(
            ','.join([ way['osm_way_id'] for way in ways ])
        )
        print('GET', url)
        resp = requests.get(url)
        if 200 != resp.status_code:
            print(resp.text)
            time.sleep(60)
            # retry
            return getSegmentClassification(ways)
        for element in resp.json()['elements']:
            if 'tags' in element:
                if 'highway' in element['tags']:
                    results.append({
                        'osm_way_id': element['id'],
                        'highway': element['tags']['highway']
                    })
            else:
                results.append({
                    'osm_way_id': element['id'],
                    'highway': None
                })
        time.sleep(2.5)
        return results
    except:
        # retry
        time.sleep(60)
        return getSegmentClassification(ways)


writer = csv.DictWriter(open('ways.csv', 'w', newline=''), fieldnames=['osm_way_id','highway'])
writer.writeheader()

with open('gbr_osm_way_ids.csv', 'r', newline='') as csvfile:
    ways = []
    reader = csv.DictReader(csvfile)
    for row in reader:
        ways.append(row)
        if 250 == len(ways):
            writer.writerows(getSegmentClassification(ways))
            ways = []
    writer.writerows(getSegmentClassification(ways))



'''

http://overpass-api.de/api/status


'''
