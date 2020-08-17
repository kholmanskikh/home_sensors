import click
from flask import current_app, g
from flask.cli import with_appcontext
from sqlalchemy.orm import sessionmaker
from sqlalchemy import create_engine

from . import model

def get_session():
    if 'db_session' not in g:
        g.db_session = current_app.config['_DB_SESSION']()

    return g.db_session

def close_session(self, e=None):
    db_session = g.pop('db_session', None)

    if db_session is not None:
        db_session.close()

def init_db():
    engine = current_app.config['_DB_ENGINE']

    model.Base.metadata.drop_all(bind=engine)
    model.Base.metadata.create_all(bind=engine, checkfirst=False)

@click.command('init-db')
@with_appcontext
def init_db_command():
    """Clear the existing data and create new tables"""

    init_db()
    click.echo('Initialized the database.')

def init_app(app):
    app.config['_DB_ENGINE'] = create_engine(app.config['DB_URI'], echo=app.config['DB_DEBUG'])
    app.config['_DB_SESSION'] = sessionmaker(bind=app.config['_DB_ENGINE'])

    app.cli.add_command(init_db_command)
    app.teardown_appcontext(close_session)

