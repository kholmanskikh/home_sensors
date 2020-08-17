from flask import Flask, render_template
from . import api
from . import db
from .validation import ValidationError
from .model import Location, Sensor, Measurement, MeasurementType
from sqlalchemy.orm import joinedload
import time


def create_app(instance_path=None):
    app = Flask(__name__, instance_path, instance_relative_config=True)

    app.config.from_pyfile('config.py', silent=False)

    db.init_app(app)

    def bad_request(msg):
        return "Bad request: '%s'" % msg, 400

    @app.errorhandler(ValidationError)
    def validation_error(e):
        return bad_request(e.args[0])
   
    app.register_blueprint(api.bp, url_prefix='/api')

    @app.route('/')
    def index():
        session = db.get_session()
        locations = session.query(Location).options(
                        joinedload(Location.sensors)).all()

        overview_data = {}
        for loc in locations:
            sensor_data = {}
            
            for sensor in loc.sensors:
                last_measurements = []
                for mtype in sensor.mtypes:
                    m = sensor.lastMeasurement(session, mtype)
                    if m is not None:
                        m = (m.mtype.name, m.value,
                                time.strftime('%H:%M %d %b %Y',
                                                time.localtime(m.timestamp))) 
                        last_measurements.append(m)
                sensor_data[sensor] = last_measurements

            overview_data[loc] = sensor_data
                
        return render_template('overview.html', overview_data=overview_data)

    return app
