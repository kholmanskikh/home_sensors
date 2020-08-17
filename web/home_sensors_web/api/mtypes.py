from flask import request, jsonify
from .. import db
from ..model import MeasurementType
from ..validation import ValidationError
from . import bp

@bp.route('/mtypes/')
def get_mtypes():
    session = db.get_session()
    mtypes = session.query(MeasurementType).all()

    return jsonify({ 'mtypes': [ m.to_json() for m in mtypes ] })

@bp.route('/mtypes/', methods=['POST'])
def new_mtype():
    name = request.json.get('name', '').strip()
    if name.strip() == '':
        raise ValidationError("'name' must be non-empty")

    session = db.get_session()
    mtype = MeasurementType(name=name)
    session.add(mtype)
    session.commit()
    
    return jsonify(mtype.to_json())

@bp.route('/mtypes/<int:id>')
def get_mtype(id):
    return jsonify(MeasurementType.from_id(db.get_session(), id).to_json())
    
