from filebeat import BaseTest
import os
import socket

"""
Test Harvesters
"""


class Test(BaseTest):

    def test_close_renamed(self):
        """
        Checks that a file is closed when its renamed / rotated
        """

        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/test.log",
            close_renamed="true",
            scan_frequency="0.1s"
        )
        os.mkdir(self.working_dir + "/log/")

        testfile1 = self.working_dir + "/log/test.log"
        testfile2 = self.working_dir + "/log/test.log.rotated"
        file = open(testfile1, 'w')

        iterations1 = 5
        for n in range(0, iterations1):
            file.write("rotation file")
            file.write("\n")

        file.close()

        filebeat = self.start_beat()

        # Let it read the file
        self.wait_until(
            lambda: self.output_has(lines=iterations1), max_timeout=10)

        os.rename(testfile1, testfile2)

        file = open(testfile1, 'w', 0)
        file.write("Hello World\n")
        file.close()

        # Wait until error shows up
        self.wait_until(
            lambda: self.log_contains(
                "Closing because close_renamed is enabled"),
            max_timeout=15)

        # Let it read the file
        self.wait_until(
            lambda: self.output_has(lines=iterations1 + 1), max_timeout=10)

        filebeat.check_kill_and_wait()

        data = self.get_registry()

        # Make sure new file was picked up. As it has the same file name,
        # one entry for the new and one for the old should exist
        assert len(data) == 2


    def test_close_removed(self):
        """
        Checks that a file is closed if removed
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/test.log",
            close_removed="true",
            scan_frequency="0.1s"
        )
        os.mkdir(self.working_dir + "/log/")

        testfile1 = self.working_dir + "/log/test.log"
        file = open(testfile1, 'w')

        iterations1 = 5
        for n in range(0, iterations1):
            file.write("rotation file")
            file.write("\n")

        file.close()

        filebeat = self.start_beat()

        # Let it read the file
        self.wait_until(
            lambda: self.output_has(lines=iterations1), max_timeout=10)

        os.remove(testfile1)

        # Wait until error shows up on windows
        self.wait_until(
            lambda: self.log_contains(
                "Closing because close_removed is enabled"),
            max_timeout=15)

        filebeat.check_kill_and_wait()

        data = self.get_registry()

        # Make sure the state for the file was persisted
        assert len(data) == 1


    def test_close_eof(self):
        """
        Checks that a file is closed if eof is reached
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/test.log",
            close_eof="true",
            scan_frequency="0.1s"
        )
        os.mkdir(self.working_dir + "/log/")

        testfile1 = self.working_dir + "/log/test.log"
        file = open(testfile1, 'w')

        iterations1 = 5
        for n in range(0, iterations1):
            file.write("rotation file")
            file.write("\n")

        file.close()

        filebeat = self.start_beat()

        # Let it read the file
        self.wait_until(
            lambda: self.output_has(lines=iterations1), max_timeout=10)


        # Wait until error shows up on windows
        self.wait_until(
            lambda: self.log_contains(
                "Closing because close_eof is enabled"),
            max_timeout=15)

        filebeat.check_kill_and_wait()

        data = self.get_registry()

        # Make sure the state for the file was persisted
        assert len(data) == 1
