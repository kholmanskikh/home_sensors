from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import Table, Column, Integer, String, Float, ForeignKey
from sqlalchemy.orm import relationship, joinedload
from .validation import ValidationError

Base = declarative_base();

_sensor_measurement_type = Table(
    'sensor_measurement_type', Base.metadata,
    Column('sensor_id', Integer, ForeignKey('sensors.id'),
            nullable=False),
    Column('mtype_id', Integer, ForeignKey('measurement_types.id'),
            nullable=False)
)

class Sensor(Base):
    __tablename__ = 'sensors'

    id = Column('id', Integer, primary_key=True)
    name = Column('name', String(50), nullable=False)
    location_id = Column('location_id', Integer, ForeignKey('locations.id'),
                        nullable=False)

    _measurements = relationship('Measurement', back_populates='sensor',
                                    cascade="all, delete-orphan" )

    location = relationship('Location', back_populates='sensors')
    mtypes = relationship('MeasurementType',
                            secondary=_sensor_measurement_type)

    def __repr__(self):
        return "<Sensor(name='%s', location='%s')>" % (self.name, self.location)

    @staticmethod
    def from_id(session, id):
        sensor = session.query(Sensor).filter_by(id=id).first()
        if sensor is None:
            raise ValidationError('Unable to find a sensor with id %d' % (id))

        return sensor

    def to_json(self):
        return { 'name': self.name, 'id': self.id, 'location_id': self.location_id,
                    'mtypes' : [ m.id for m in self.mtypes ]}

    def _validateMtype(self, mtype):
        if mtype not in self.mtypes:
            raise ValidationError('Sensor does not support the %s type' %
                                    (mtype.name))
        
    def addMeasurement(self, session, measurement):
        self._validateMtype(measurement.mtype)

        measurement.sensor = self

        session.add(measurement)

    def queryMeasurements(self, session, mtype=None):
        self._validateMtype(mtype)

        query = session.query(Measurement).options(joinedload(Measurement.mtype)).\
                    filter_by(sensor=self)

        if mtype is not None:
            return query.filter_by(mtype=mtype)
        else:
            return query

    def lastMeasurement(self, session, mtype):
        return self.queryMeasurements(session, mtype).\
                order_by(Measurement.timestamp.desc()).first()

class Location(Base):
    __tablename__ = 'locations'

    id = Column('id', Integer, primary_key=True)
    name = Column('name', String(50), nullable=False)

    sensors = relationship('Sensor', back_populates='location')

    def __repr__(self):
        return "<Location(name='%s')>" % (self.name)

    @staticmethod
    def from_id(session, id):
        location = session.query(Location).filter_by(id=id).first()
        if location is None:
            raise ValidationError('Unable to find a location with id %d' % (id))

        return location

    def to_json(self):
        return { 'id': self.id, 'name': self.name }

class MeasurementType(Base):
    __tablename__ = 'measurement_types'

    id = Column('id', Integer, primary_key=True)
    name = Column('name', String(50), nullable=False)

    def __repr__(self):
        return "<MeasurementType(name='%s')>" % (self.name)

    @staticmethod
    def from_id(session, id):
        mtype = session.query(MeasurementType).filter_by(id=id).first()
        if mtype is None:
            raise ValidationError('Unable to find a type with id %d' % (id))

        return mtype

    def to_json(self):
        return { 'id': self.id, 'name': self.name }

class Measurement(Base):
    __tablename__ = 'measurements'

    id = Column('id', Integer, primary_key=True)
    mtype_id = Column('mtype_id', Integer, ForeignKey('measurement_types.id'),
                        nullable=False)
    sensor_id = Column('sensor_id', Integer, ForeignKey('sensors.id'),
                        nullable=False)
    timestamp = Column('timestamp', Integer, nullable=False)
    value = Column('value', Float, nullable=False)

    mtype = relationship('MeasurementType')
    sensor = relationship('Sensor')

    def __repr__(self):
        return "<Measurement(mtype='%s', timestamp='%s', value='%s')>" % (
                self.mtype, self.timestamp, self.value)

    @staticmethod
    def from_id(session, id):
        m = session.query(Measurement).filter_by(id=id).first()
        if m is None:
            raise ValidationError('Unable to find a measurement with id %d' % 
                                    (id))

        return m

    def to_json(self):
        return { 'id': self.id, 'mtype_id': self.mtype_id, 'sensor_id': self.sensor_id,
                'timestamp': self.timestamp, 'value': self.value}

