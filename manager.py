import os

from cpbox.app.devops import DevOpsApp
from cpbox.tool import dockerutil
from cpbox.tool import file
from cpbox.tool import functocli
from cpbox.tool import template
from cpbox.tool import utils

APP_NAME = 'cc-robot'


class App(DevOpsApp):

    def __init__(self, **kwargs):
        DevOpsApp.__init__(self, APP_NAME, **kwargs)

    def run(self):
        self.shell_run('cd app && go run main.go')

    def check_run(self):
        self.shell_run('curl 0:3333/check-health')


if __name__ == '__main__':
    common_args_option = {
        'args': [],
        'default_values': {
        }
    }
    functocli.run_app(App, common_args_option=common_args_option)
