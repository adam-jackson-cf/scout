import yaml


def unsafe(data):
    return yaml.load(data, Loader=yaml.Loader)
