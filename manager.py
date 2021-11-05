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
        self.dev_screen_name = '%s-%s' % (APP_NAME, 'dev')
        self.prod_container = '%s-%s' % (APP_NAME, 'prod')

    def run(self):
        self._prepare_runtime('dev')
        self._restart_dev()

    def restart_app_container(self):
        self.stop_container(self.prod_container, timeout=1)
        self.remove_container(self.prod_container, force=True)

        volumes = {
            '/etc/resolv.conf': '/etc/resolv.conf',
        }

        args = dockerutil.base_docker_args(container_name=self.prod_container, volumes=volumes)
        cmd_data = {'image': docker_image, 'args': args}
        cmd = template.render_str('docker run -d --restart always {{ args }} {{ image }}', cmd_data)
        self.shell_run(cmd)

    def check_run(self):
        self.shell_run('curl 0:3333/check-health')

    def build(self):
        if not sys.platform.startswith('linux'):
            self.logger.error('build must be on linux system')
            return

        self._prepare_runtime('prod')
        self.shell_run('cd app && go build main.go')
        self.shell_run('docker build -t %s .' % docker_image)

    def _restart_dev(self):
        self.stop_dev()
        self.shell_run('cd app && screen -L -Logfile %s -S %s -dm go run main.go -env=dev' % (self.app_logs_dir + '/app-dev.log', self.dev_screen_name))
        self.shell_run('screen -S %s -X colon "logfile flush 0^M"' % self.dev_screen_name)

    def stop_dev(self):
        self.shell_run('screen -ls | grep %s | cut -d . -f 1 | xargs pkill -TERM -P ' % self.dev_screen_name, exit_on_error=False)

    def _prepare_runtime(self, env):
        self._prepare_runtime_config('api.yaml', env)

    def _prepare_runtime_config(self, name, env):
        names = name.split('.')
        self.shell_run('cp app/config/%s.%s.%s app/config/%s' % (names[0], env, names[1], name))


if __name__ == '__main__':
    common_args_option = {
        'args': [],
        'default_values': {
        }
    }
    functocli.run_app(App, common_args_option=common_args_option)
