import json
import time
from engine_fixture import EngineFixture


class BrakeIntegration(EngineFixture):
    def test_gate_precedes_every_engine_on_path(self):
        self.start()
        self.transition("ready-to-drive")
        self.assertFalse(self.sounding())
        for command in ("engine:on", "engine:vol:65", "engine:style:sport"):
            self.redis("PUBLISH", "music", command)
        self.wait(lambda: json.loads(self.config.read_text())["engine_sound"]["style"] == "sport")
        self.redis("HSET", "vehicle", "brake:left", "on")
        self.redis("PUBLISH", "vehicle", "brake:left")
        self.redis("PUBLISH", "vehicle", "state")
        time.sleep(0.1)
        self.assertFalse(self.sounding())
        self.assertEqual(self.on_frames(), 0)
        self.transition("parked", "on")
        self.transition("ready-to-drive", "on")
        self.wait(lambda: "id=0x4EA data=01 41 02" in self.log.read_text())
        self.assertTrue(self.sounding())

    def test_release_latches_but_restart_and_next_interval_do_not(self):
        self.start()
        self.transition("ready-to-drive", "on")
        self.wait(lambda: self.on_frames() > 0)
        stamp = self.config.stat().st_mtime_ns
        self.redis("PUBLISH", "music", "settings:engine.require_left_brake:true")
        self.wait(lambda: self.config.stat().st_mtime_ns != stamp)
        self.assertTrue(self.sounding())
        self.redis("HSET", "vehicle", "brake:left", "off")
        self.redis("PUBLISH", "vehicle", "brake:left")
        self.redis("PUBLISH", "vehicle", "state")
        time.sleep(0.1)
        self.assertTrue(self.sounding())
        self.stop()
        count = self.on_frames()
        self.start()
        self.assertFalse(self.sounding())
        self.assertEqual(self.on_frames(), count)
        self.transition("parked")
        self.redis("HDEL", "vehicle", "brake:left")
        self.redis("HSET", "vehicle", "state", "ready-to-drive")
        self.redis("PUBLISH", "vehicle", "state")
        self.wait(lambda: self.redis("HGET", "music", "engine_vehicle_state") == "ready-to-drive")
        self.assertFalse(self.sounding())
        self.assertEqual(self.on_frames(), count)

    def test_disabled_gate_keeps_existing_activation(self):
        config = json.loads(self.config.read_text())
        config["engine_sound"].pop("require_left_brake")
        self.config.write_text(json.dumps(config))
        self.start()
        self.transition("ready-to-drive")
        self.wait(lambda: self.on_frames() > 0)
        self.assertTrue(self.sounding())
