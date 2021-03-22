import json
import urllib.request
import urllib.parse
import logging
import argparse
import zmq

# Relative to the base api url
API_URL_MTYPES = 'mtypes/'
API_URL_MEASUREMENTS = 'measurements/'
API_URL_SENSORS = 'sensors/'

def read_json_from_url(url):
    reply = urllib.request.urlopen(url)

    if reply.status != 200:
        raise RuntimeError("'%s' returned HTTP code %d, not 200" % (url, reply.status))

    return json.loads(reply.read().decode('utf-8'))

def get_obj_with_field_value(collection, field_name, field_value):
    for obj in collection:
        if field_name in obj and obj[field_name] == field_value:
            return obj

    return None

def post_measurement(url, m):
    jm = json.dumps(m)

    request = urllib.request.Request(url,
                                    data=jm.encode('utf-8'),
                                    headers={'Content-Type': 'application/json'},
                                    method='POST')

    reply = urllib.request.urlopen(request)
    if reply.status != 200:
        raise RuntimeError("POST '%s' to '%s' returned '%s': %s" % 
                            (jm, url, reply.status, reply.read()))

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Read from ZMQ and publish to API')
    parser.add_argument('-z', dest='zmq_endpoint',
                        help='ZMQ endpoint to subscribe to',
                        required=True)
    parser.add_argument('-a', dest='base_api_url',
                        help='Base API URL',
                        required=True)
    parser.add_argument('-d', dest='debugging',
                        action='store_true',
                        help='Enable debugging')

    args = parser.parse_args()

    if args.debugging:
        logging.basicConfig(level=logging.DEBUG) 
    else:
        logging.basicConfig(level=logging.INFO)

    zmq_endpoint = args.zmq_endpoint

    api_url_base = args.base_api_url
    api_url_mtypes = urllib.parse.urljoin(api_url_base, API_URL_MTYPES)
    api_url_measurements= urllib.parse.urljoin(api_url_base, API_URL_MEASUREMENTS)
    api_url_sensors = urllib.parse.urljoin(api_url_base, API_URL_SENSORS)

    logging.info("Working with the API on '%s'" % (api_url_base))

    mtypes = read_json_from_url(api_url_mtypes)['mtypes']
    mtype_names = tuple(m['name'] for m in mtypes)
    logging.info('API supports %d measurement types: %s' % (len(mtype_names), ', '.join(sorted(mtype_names))))

    sensors = read_json_from_url(api_url_sensors)['sensors']
    logging.info('API supports %d sensors' % (len(sensors)))

    context = zmq.Context()
    socket = context.socket(zmq.SUB)
    socket.connect(zmq_endpoint)
    socket.setsockopt(zmq.SUBSCRIBE, b'')

    logging.info("Receiving messages from the '%s' ZMQ endpoint" % (zmq_endpoint))
    logging.info("Begin publishing messages with measurement data to the API")

    while True:
        message = socket.recv_json()

        logging.debug("Received message: %s" % message)

        device_id = message.pop('device_id')
        timestamp = message.pop('timestamp')
        message_type = message['type']

        if message_type == 'Error':
            logging.debug('Skipping, since this is an error message')
            continue

        mtype = get_obj_with_field_value(mtypes, 'name', message_type)
        if mtype is None:
            logging.info("API does not support the '%s' measurement type" % (message_type))
            continue

        sensor = get_obj_with_field_value(sensors, 'id', device_id)
        if sensor is None:
            logging.info("API does not have a sensor with id '%s'" % (device_id))
            continue


        measurement = {}
        measurement['mtype_id'] = mtype['id']
        measurement['sensor_id'] = sensor['id']
        measurement['timestamp'] = timestamp
        measurement['value'] = message.pop('value')

        post_measurement(api_url_measurements, measurement)

