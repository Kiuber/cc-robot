import os
import sys

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
        self._run_for_debug()

    def run_as_service(self):
        self.shell_run('docker run -d --name %s %s' % (APP_NAME, APP_NAME))

    def check_run(self):
        self.shell_run('curl 0:3333/check-health')

    def build(self):
        if not sys.platform.startswith('linux'):
            self.logger.error('build must be on linux system')
            return

        self.shell_run('cd app && go build main.go')
        self.shell_run('docker build -t %s .' % APP_NAME)

    def _run_for_debug(self):
        self.shell_run('cd app && go run main.go')


if __name__ == '__main__':
    common_args_option = {
        'args': [],
        'default_values': {
        }
    }
    functocli.run_app(App, common_args_option=common_args_option)
