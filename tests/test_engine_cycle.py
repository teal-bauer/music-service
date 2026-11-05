import json
import subprocess
import time
from engine_fixture import EngineFixture


class CycleIntegration(EngineFixture):
    def configure(self, **settings):
        config = json.loads(self.config.read_text())
        config["engine_sound"].update(settings)
        self.config.write_text(json.dumps(config))

    def subscribe(self):
        self.messages = self.root / "notifications"
        with self.messages.open("w") as output:
            self.subscription = subprocess.Popen([
                "redis-cli", "-p", str(self.port), "--raw", "SUBSCRIBE", "scootui:notification"],
                stdout=output, stderr=subprocess.DEVNULL)
        self.addCleanup(self.terminate, self.subscription)
        self.wait(lambda: "scootui:notification" in self.messages.read_text())

    def notifications(self):
        return [json.loads(line) for line in self.messages.read_text().splitlines() if line.startswith("{")]

    def style(self):
        return json.loads(self.config.read_text())["engine_sound"]["style"]

    def test_rotation_selects_first_frame_and_notifies_without_echo(self):
        self.configure(style="electric", cycle_on_drive=True, notify_on_drive=True)
        self.start()
        self.subscribe()
        start = time.monotonic()
        self.transition("ready-to-drive")
        self.wait(lambda: len(self.notifications()) == 1)
        self.assertLess(time.monotonic() - start, 2)
        self.wait(lambda: "id=0x4EA data=01 32 00" in self.log.read_text())
        self.wait(lambda: self.style() == "v8")
        message = self.notifications()[0]
        self.assertEqual(message["title"], "Engine sound: V8")
        self.assertNotIn("volume", json.dumps(message))
        self.redis("PUBLISH", "vehicle", "state")
        time.sleep(0.1)
        self.assertEqual(len(self.notifications()), 1)
        self.assertEqual(self.style(), "v8")
        self.transition("parked")
        self.transition("ready-to-drive")
        self.wait(lambda: len(self.notifications()) == 2)
        self.assertEqual(self.notifications()[1]["title"], "Engine sound: Moped")
        self.stop()
        self.start()
        time.sleep(0.1)
        self.assertEqual(self.style(), "moped")
        self.assertEqual(len(self.notifications()), 2)

    def test_notification_can_be_disabled_without_disabling_cycle(self):
        self.configure(cycle_on_drive=True, notify_on_drive=False)
        self.start()
        self.subscribe()
        self.transition("ready-to-drive")
        self.wait(lambda: self.style() == "moped")
        time.sleep(0.1)
        self.assertEqual(self.notifications(), [])

    def test_fixed_style_can_notify_and_off_mode_stays_off(self):
        self.configure(enabled=False, cycle_on_drive=False, notify_on_drive=True)
        self.start()
        self.subscribe()
        self.transition("ready-to-drive")
        self.wait(lambda: len(self.notifications()) == 1)
        self.assertEqual(self.notifications()[0]["title"], "Engine sound off")
        self.assertEqual(self.style(), "v8")
        self.assertEqual(self.on_frames(), 0)
        self.assertFalse(self.sounding())

    def test_silent_cycle_does_not_enable_engine_sound(self):
        self.configure(enabled=False, cycle_on_drive=True, notify_on_drive=True)
        self.start()
        self.subscribe()
        self.transition("ready-to-drive")
        self.wait(lambda: self.style() == "moped" and len(self.notifications()) == 1)
        self.assertEqual(self.notifications()[0]["title"], "Engine sound off")
        self.assertEqual(self.on_frames(), 0)

    def test_cycle_combines_with_optional_brake_admission(self):
        self.configure(require_left_brake=True, cycle_on_drive=True, notify_on_drive=True)
        self.start()
        if self.redis("HGET", "music", "engine_brake_required") != "true":
            self.skipTest("brake admission is not available in this branch")
        self.subscribe()
        self.transition("ready-to-drive", "off")
        self.wait(lambda: self.style() == "moped" and len(self.notifications()) == 1)
        self.assertEqual(self.notifications()[0]["title"], "Engine sound off")
        self.assertEqual(self.on_frames(), 0)
        self.transition("parked")
        self.transition("ready-to-drive", "on")
        self.wait(lambda: len(self.notifications()) == 2)
        self.assertEqual(self.notifications()[1]["title"], "Engine sound: Sport")
        self.wait(lambda: "id=0x4EA data=01 32 02" in self.log.read_text())

    def test_both_features_default_off(self):
        self.start()
        self.subscribe()
        self.transition("ready-to-drive")
        self.wait(lambda: self.on_frames() > 0)
        time.sleep(0.1)
        self.assertEqual(self.style(), "v8")
        self.assertEqual(self.notifications(), [])
