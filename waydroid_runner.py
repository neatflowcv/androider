#!/usr/bin/env python3
# Copyright 2021 Oliver Smith
# SPDX-License-Identifier: GPL-3.0-or-later
# Python wrapper for Go CLI - receives JSON args and calls waydroid tools

import json
import sys
import os
import argparse
import logging
import traceback
import dbus.mainloop.glib
import dbus
import dbus.exceptions

# Add tools directory to path so we can import the modules
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'tools'))

from tools import actions
from tools import config
from tools import helpers
from tools.helpers import logging as tools_logging


class MockArgs(argparse.Namespace):
    """Mock argparse.Namespace to simulate the args structure from argparse"""
    def __init__(self, data):
        super().__init__()
        
        # Set default values first
        self.cache = {}
        self.work = config.defaults["work"]
        self.config = self.work + "/waydroid.cfg"
        self.log = self.work + "/waydroid.log"
        self.sudo_timer = True
        self.timeout = 1800
        
        # Set values from JSON data
        for key, value in data.items():
            setattr(self, key, value)
            
        # Handle extra_args if present
        if hasattr(self, 'extra_args') and self.extra_args:
            for key, value in self.extra_args.items():
                setattr(self, key, value)
            delattr(self, 'extra_args')
            
        # Ensure required attributes exist with defaults
        if not hasattr(self, 'action'):
            self.action = None
        if not hasattr(self, 'subaction'):
            self.subaction = None
        if not hasattr(self, 'log'):
            self.log = self.work + "/waydroid.log"
        if not hasattr(self, 'details_to_stdout'):
            self.details_to_stdout = False
        if not hasattr(self, 'verbose'):
            self.verbose = False
        if not hasattr(self, 'quiet'):
            self.quiet = False
        if not hasattr(self, 'wait_for_init'):
            self.wait_for_init = False


def main_with_args(args_data):
    """Main function that takes pre-parsed args data instead of parsing argv"""
    
    def actionNeedRoot(action):
        if os.geteuid() != 0:
            raise RuntimeError(
                "Action \"{}\" needs root access".format(action))

    # Wrap everything to display nice error messages
    args = None
    try:
        # Create args from JSON data instead of parsing command line
        args = MockArgs(args_data)
        
        # Set up log file path
        if os.geteuid() == 0:
            if not os.path.exists(args.work):
                os.mkdir(args.work)
        elif not os.path.exists(args.log):
            args.log = "/tmp/tools.log"

        tools_logging.init(args)

        dbus.mainloop.glib.DBusGMainLoop(set_as_default=True)
        dbus.mainloop.glib.threads_init()
        dbus_name_scope = None

        if not actions.initializer.is_initialized(args) and \
                args.action and args.action not in ("init", "first-launch", "log"):
            if args.wait_for_init:
                try:
                    dbus_name_scope = dbus.service.BusName("id.waydro.Container", dbus.SystemBus(), do_not_queue=True)
                    actions.wait_for_init(args)
                except dbus.exceptions.NameExistsException:
                    print('ERROR: WayDroid service is already awaiting initialization')
                    return 1
            else:
                print('ERROR: WayDroid is not initialized, run "waydroid init"')
                return 0

        # Initialize or require config
        if args.action == "init":
            actionNeedRoot(args.action)
            actions.init(args)
        elif args.action == "upgrade":
            actionNeedRoot(args.action)
            actions.upgrade(args)
        elif args.action == "session":
            if args.subaction == "start":
                actions.session_manager.start(args)
            elif args.subaction == "stop":
                actions.session_manager.stop(args)
            else:
                logging.info(
                    "Run waydroid {} -h for usage information.".format(args.action))
        elif args.action == "container":
            actionNeedRoot(args.action)
            if args.subaction == "start":
                if dbus_name_scope is None:
                    try:
                        dbus_name_scope = dbus.service.BusName("id.waydro.Container", dbus.SystemBus(), do_not_queue=True)
                    except dbus.exceptions.NameExistsException:
                        print('ERROR: WayDroid container service is already running')
                        return 1
                actions.container_manager.start(args)
            elif args.subaction == "stop":
                actions.container_manager.stop(args)
            elif args.subaction == "restart":
                actions.container_manager.restart(args)
            elif args.subaction == "freeze":
                actions.container_manager.freeze(args)
            elif args.subaction == "unfreeze":
                actions.container_manager.unfreeze(args)
            else:
                logging.info(
                    "Run waydroid {} -h for usage information.".format(args.action))
        elif args.action == "app":
            if args.subaction == "install":
                actions.app_manager.install(args)
            elif args.subaction == "remove":
                actions.app_manager.remove(args)
            elif args.subaction == "launch":
                actions.app_manager.launch(args)
            elif args.subaction == "intent":
                actions.app_manager.intent(args)
            elif args.subaction == "list":
                actions.app_manager.list(args)
            else:
                logging.info(
                    "Run waydroid {} -h for usage information.".format(args.action))
        elif args.action == "prop":
            if args.subaction == "get":
                actions.prop.get(args)
            elif args.subaction == "set":
                actions.prop.set(args)
            else:
                logging.info(
                    "Run waydroid {} -h for usage information.".format(args.action))
        elif args.action == "shell":
            actionNeedRoot(args.action)
            helpers.lxc.shell(args)
        elif args.action == "logcat":
            actionNeedRoot(args.action)
            helpers.lxc.logcat(args)
        elif args.action == "show-full-ui":
            actions.app_manager.showFullUI(args)
        elif args.action == "first-launch":
            actions.remote_init_client(args)
            if actions.initializer.is_initialized(args):
                actions.app_manager.showFullUI(args)
        elif args.action == "status":
            actions.status.print_status(args)
        elif args.action == "adb":
            if args.subaction == "connect":
                helpers.net.adb_connect(args)
            elif args.subaction == "disconnect":
                helpers.net.adb_disconnect(args)
            else:
                logging.info("Run waydroid {} -h for usage information.".format(args.action))
        elif args.action == "log":
            if hasattr(args, 'clear_log') and args.clear_log:
                helpers.run.user(args, ["truncate", "-s", "0", args.log])
            try:
                lines = getattr(args, 'lines', '60')
                helpers.run.user(
                    args, ["tail", "-n", lines, "-F", args.log], output="tui")
            except KeyboardInterrupt:
                pass
        else:
            logging.info("Run waydroid -h for usage information.")

        #logging.info("Done")
        return 0

    except Exception as e:
        # Dump log to stdout when args (and therefore logging) init failed
        if not args:
            logging.getLogger().setLevel(logging.DEBUG)

        logging.info("ERROR: " + str(e))
        logging.info("See also: <https://github.com/waydroid>")
        logging.debug(traceback.format_exc())

        # Hints about the log file (print to stdout only)
        log_hint = "Run 'waydroid log' for details."
        if not args or not os.path.exists(args.log):
            log_hint += (" Alternatively you can use '--details-to-stdout' to"
                         " get more output, e.g. 'waydroid"
                         " --details-to-stdout init'.")
        print(log_hint)
        return 1


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: {} <json_args>".format(sys.argv[0]), file=sys.stderr)
        sys.exit(1)
    
    try:
        # Parse JSON args from command line
        args_json = sys.argv[1]
        args_data = json.loads(args_json)
        
        # Set umask as in original waydroid.py
        os.umask(0o0022)
        
        # Call main function with parsed args
        exit_code = main_with_args(args_data)
        sys.exit(exit_code)
        
    except json.JSONDecodeError as e:
        print("Error parsing JSON args: {}".format(e), file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print("Error: {}".format(e), file=sys.stderr)
        sys.exit(1)
