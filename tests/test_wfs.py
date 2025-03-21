import sys
import requests

domain = "http://localhost:5555"

if 1 < len(sys.argv):
    domain = sys.argv[1]

url = "{0}/api/v1/layer/tl_2017_us_zcta510/wfs".format(domain)

data = {
    "method": "get_feature",
    "limit": 1,
    "layer": "tl_2017_us_zcta510",
    "filters": [
        {
            "test": "contains",
            "wkt": "POINT(-105.02929687500001 48.019324184801185)"
        }
    ]
}

resp = requests.post(url, json=data)

print(resp.text)
