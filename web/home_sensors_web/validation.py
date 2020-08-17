class ValidationError(ValueError):
    pass


def validate_type(type_, value):
    try:
        res = type_(value)
    except (TypeError, ValueError):
        raise ValidationError("Unable to parse '%s' from '%s'" %
                                (type_.__name__, value))

    return res
