import sys
import requests

domain = "http://localhost:5555"

if 1 < len(sys.argv):
    domain = sys.argv[1]

url = "{0}/api/v1/layer/tl_2017_us_zcta510/tile/0/0/0.mvt".format(domain)

try:
    resp = requests.get(url, timeout=0.00001)
except Exception as e:
    print(e)
