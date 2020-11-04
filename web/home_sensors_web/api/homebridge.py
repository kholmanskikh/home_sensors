from flask import request, jsonify
from .. import db
from ..model import Measurement, MeasurementType, Sensor
from ..validation import ValidationError
from . import bp

def get_latest_measurement(session, sensor, name):
    mtype = session.query(MeasurementType).filter_by(name=name).first()
    if mtype is None:
        raise ValidationError("No '%s' mtype defined" % (name))

    if mtype not in sensor.mtypes:
        raise ValidationError("Sensor %d does not support '%s'" % (sensor.id, name))

    measurement = session.query(Measurement).filter_by(sensor_id=sensor.id).\
                    filter_by(mtype_id=mtype.id).\
                    order_by(Measurement.timestamp.desc()).first()

    if measurement is None:
        raise ValidationError("No '%s' measurements for sensor %d" % (
                                name, sensor.id))

    return measurement.value

@bp.route('/homebridge/<int:sensor_id>')
def get_latest_temphumid(sensor_id):
    session = db.get_session()

    sensor = Sensor.from_id(session, sensor_id)

    return jsonify({'temperature': get_latest_measurement(session, sensor, 'Temperature'),
                    'humidity': get_latest_measurement(session, sensor, 'Humidity')})
