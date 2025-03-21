import sys
import json
import requests
import unittest


domain = "http://localhost:5555"

if 1 < len(sys.argv):
    domain = sys.argv[1]


class TestLayers(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        cls.layers = []

    # @classmethod
    # def tearDownClass(cls):
    #     cls.layers = []

    # def setUp(self):
    # def tearDown(self):

    def test_getLayers(self):
        # get layers from server
        resp = requests.get("{0}/api/v1/layers".format(domain))
        self.assertEqual(200, resp.status_code)
        self.assertNotEqual(0, len(resp.content))

        result = resp.json()
        self.assertEqual("ok", result['status'])

        # add layers to layer
        for layer in result['data']['layers']:
            self.layers.append(layer)

    def test_getLayer_fail(self):
        # get the layer
        resp = requests.get("{0}/api/v1/layer/{1}".format(domain, "FAIL"))
        self.assertEqual(404, resp.status_code)
        self.assertNotEqual(0, len(resp.content))

        result = resp.json()
        self.assertEqual("error", result['status'])

    def test_getLayer_success(self):
        # check if there are layers
        self.assertNotEqual(0, len(self.layers))

        # check each layer
        for layer1 in self.layers:

            # get the layer
            resp = requests.get("{0}/api/v1/layer/{1}".format(domain, layer1['layer_name']))
            self.assertEqual(200, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

            result = resp.json()
            self.assertEqual("ok", result['status'])

            layer2 = result['data']['layer']

            # check each attribute of the layer
            for i in layer1:
                self.assertEqual(layer1[i], layer1[i])

    def test_tile_fail(self):
        # check if there are layers
        self.assertNotEqual(0, len(self.layers))

        # check each layer
        for layer in self.layers:

            # get a tile with bad zoom
            resp = requests.get("{0}/api/v1/layer/{1}/tile/-1/0/0.mvt".format(domain, layer['layer_name']))
            self.assertEqual(404, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

            # get tile with bad filter
            resp = requests.get('{0}/api/v1/layer/{1}/tile/0/0/0.mvt'.format(domain, layer['layer_name']), params={
                'filters': json.dumps({"conditions":[{"method":"FAIL"}]})
            })

            self.assertEqual(400, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

            result = resp.json()
            self.assertEqual("error", result['status'])

    def test_tile_success(self):
        # check if there are layers
        self.assertNotEqual(0, len(self.layers))

        # check each layer
        for layer in self.layers:

            # get a tile for layer
            resp = requests.get("{0}/api/v1/layer/{1}/tile/0/0/0.mvt".format(domain, layer['layer_name']))
            self.assertEqual(200, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

    def test_wfs_fail(self):
        # check if there are layers
        self.assertNotEqual(0, len(self.layers))

        # check each layer
        for layer in self.layers:

            # get a tile for layer
            resp = requests.post("{0}/api/v1/layer/{1}/wfs".format(domain, layer['layer_name']))

            self.assertNotEqual(200, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

            result = resp.json()
            self.assertEqual("error", result['status'])

    def test_wfs_noFilter_success(self):
        # check if there are layers
        self.assertNotEqual(0, len(self.layers))

        # check each layer
        for layer in self.layers:

            # get a tile for layer
            resp = requests.post("{0}/api/v1/layer/{1}/wfs".format(domain, layer['layer_name']), json={
                "method": "get_feature",
                "limit": 1,
                "layer": layer['layer_name']
            })

            self.assertEqual(200, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

            result = resp.json()
            self.assertEqual("ok", result['status'])
            self.assertNotEqual(0, len(result['data']['geojson']['features']))

    def test_wfs_withFilter_success(self):
        # check if there are layers
        self.assertNotEqual(0, len(self.layers))

        # check each layer
        for layer in self.layers:

            # get a tile for layer
            resp = requests.post("{0}/api/v1/layer/{1}/wfs".format(domain, layer['layer_name']), json={
                "method": "get_feature",
                "limit": 1,
                "layer": layer['layer_name'],
                "filters": [
                    {
                        "test": "contains",
                        "wkt": "POINT(-105.02929687500001 48.019324184801185)"
                    }
                ]
            })

            self.assertEqual(200, resp.status_code)
            self.assertNotEqual(0, len(resp.content))

            result = resp.json()
            self.assertEqual("ok", result['status'])


if __name__ == '__main__':
    # unittest.main()
    # print(dir(unittest.TestLoader))
    # unittest.main(verbosity=2)

    # set test order
    suite = unittest.TestSuite()
    for test in [
                    'test_getLayers',
                    'test_getLayer_fail',
                    'test_getLayer_success',
                    'test_tile_fail',
                    'test_tile_success',
                    'test_wfs_fail',
                    'test_wfs_noFilter_success',
                    'test_wfs_withFilter_success'
                ]:
        suite.addTest(TestLayers(test))

    # suite.addTest(TestLayers('test_get_layers'))
    # suite.addTest(TestLayers('test_get_layer'))
    # suite.addTest(TestLayers('test_get_tile'))

    runner = unittest.TextTestRunner(failfast=True, verbosity=2)
    runner.run(suite)
