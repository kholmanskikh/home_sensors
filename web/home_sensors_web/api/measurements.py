from flask import request, jsonify
import time
from .. import db
from ..model import Measurement, MeasurementType, Sensor
from ..validation import ValidationError, validate_type
from . import bp

@bp.route('/measurements/')
def get_measurements():
    measurements = db.get_session().query(Measurement)

    sensor_id = request.args.get('sensor_id', None)
    if sensor_id is not None:
        sensor_id = validate_type(int, sensor_id)
        measurements = measurements.filter_by(sensor_id=sensor_id)

    mtype_id = request.args.get('mtype_id', None)
    if mtype_id is not None:
        mtype_id = validate_type(int, mtype_id)
        measurements = measurements.filter_by(mtype_id=mtype_id)

    sort_by = request.args.get('sort_by', None)
    if sort_by is not None:
        try:
            field, order = sort_by.split('.', 1)
        except ValueError:
            raise ValidationError("Invalid format of the sort_by field '%s'"
                                    % (sort_by))
        if field == 'timestamp':
            column = Measurement.timestamp
        else:
            raise ValidationError("Unsupported sort_by field '%s'" % (field))

        if order == 'asc':
            column = column.asc()
        elif order == 'desc':
            column = column.desc()
        else:
            raise ValidationError("Unsupported order '%s'" % (order))

        measurements = measurements.order_by(column)

    start = request.args.get('start', None)
    if start is not None:
        start = validate_type(int, start)
        if start < 0:
            raise ValidationError("'start' must be non-negative")

        measurements = measurements.offset(start)

    limit = request.args.get('limit', None)
    if limit is not None:
        limit = validate_type(int, limit)
        if limit <= 0:
            raise ValidationError("'limit' must be positive")
    else:
        limit = 25

    measurements = measurements.limit(limit)

    return jsonify({ 'measurements': [ m.to_json() for m in measurements.all() ] })

@bp.route('/measurements/', methods=['POST'])
def new_measurement():
    sensor_id = validate_type(int, request.json.get('sensor_id'))
    mtype_id = validate_type(int, request.json.get('mtype_id'))
    value = validate_type(float, request.json.get('value'))

    timestamp = request.json.get('timestamp')
    if timestamp is None:
        timestamp = int(time.time())
    else:
        timestamp = validate_type(int, timestamp)

    session = db.get_session()

    m = Measurement()
    m.mtype = MeasurementType.from_id(session, mtype_id)
    m.value = value
    m.timestamp = timestamp

    sensor = Sensor.from_id(session, sensor_id)
    sensor.addMeasurement(session, m)

    session.commit()

    return jsonify(m.to_json())

@bp.route('/measurements/<int:id>')
def get_measurement(id):
    return jsonify(Measurement.from_id(db.get_session(), id).to_json())
