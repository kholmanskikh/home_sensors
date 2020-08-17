from flask import request, jsonify
from .. import db
from ..model import Location
from ..validation import ValidationError
from . import bp

@bp.route('/locations/')
def get_locations():
    session = db.get_session()
    locations = session.query(Location).all()

    return jsonify({'locations': [ l.to_json() for l in locations]})


@bp.route('/locations/', methods=['POST'])
def new_location():
    name = request.json.get('name', '').strip()
    if name == '':
        raise ValidationError("'name' must be non-empty")

    location = Location(name=name.strip())

    session = db.get_session()
    session.add(location)
    session.commit()

    return jsonify(location.to_json())

@bp.route('/locations/<int:id>')
def get_location(id):
    return jsonify(Location.from_id(db.get_session(), id).to_json())
