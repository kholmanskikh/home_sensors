from flask import Blueprint

bp = Blueprint('api', '__name__')

from . import locations, sensors, mtypes, measurements, homebridge

