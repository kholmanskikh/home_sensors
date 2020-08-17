from flask import request, jsonify
from .. import db
from ..model import Sensor, Location, MeasurementType
from ..validation import ValidationError, validate_type
from . import bp

@bp.route('/sensors/')
def get_sensors():
    session = db.get_session()
    sensors = session.query(Sensor).all()

    return jsonify({ 'sensors': [ s.to_json() for s in sensors ] })

@bp.route('/sensors/', methods=['POST'])
def new_sensor():
    name = request.json.get('name', '').strip()
    if name == '':
        raise ValidationError("'name' must be non-empty")

    location_id = validate_type(int, request.json.get('location_id'))

    arr = request.json.get('mtype_ids')
    if not isinstance(arr, list):
        raise ValidationError("'mtype_ids' should be an array")
    mtype_ids = []
    for i in arr:
        mtype_ids.append(validate_type(int, i))

    session = db.get_session()
    
    sensor = Sensor(name=name)
    sensor.location = Location.from_id(session, location_id)
    for mtype_id in mtype_ids:
        sensor.mtypes.append(MeasurementType.from_id(session, mtype_id))

    session.add(sensor)
    session.commit()

    return jsonify(sensor.to_json())

@bp.route('/sensors/<int:id>')
def get_sensor(id):
    return jsonify(Sensor.from_id(db.get_session(), id).to_json())
