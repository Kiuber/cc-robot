import os
import sys

from cpbox.app.devops import DevOpsApp
from cpbox.tool import dockerutil
from cpbox.tool import file
from cpbox.tool import functocli
from cpbox.tool import template
from cpbox.tool import utils

APP_NAME = 'cc-robot'
docker_image = APP_NAME


class App(DevOpsApp):

    def __init__(self, **kwargs):
        DevOpsApp.__init__(self, APP_NAME, **kwargs)

    def run(self):
        self._run_for_dev()

    def restart_app_container(self):
        container = '%s-%s' % (APP_NAME, 'prod')
        self.stop_container(container, timeout=1)
        self.remove_container(container, force=True)

        volumes = {
            '/etc/resolv.conf': '/etc/resolv.conf',
        }

        args = dockerutil.base_docker_args(container_name=container, volumes=volumes)
        cmd_data = {'image': docker_image, 'args': args}
        cmd = template.render_str('docker run -d --restart always {{ args }} {{ image }}', cmd_data)
        self.shell_run(cmd)

    def check_run(self):
        self.shell_run('curl 0:3333/check-health')

    def build(self):
        if not sys.platform.startswith('linux'):
            self.logger.error('build must be on linux system')
            return

        self.shell_run('cd app && go build main.go')
        self.shell_run('docker build -t %s .' % docker_image)

    def _run_for_dev(self):
        self.shell_run('cd app && go run main.go -env=dev')


if __name__ == '__main__':
    common_args_option = {
        'args': [],
        'default_values': {
        }
    }
    functocli.run_app(App, common_args_option=common_args_option)
