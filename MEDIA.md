# Song and sound assets

The source-code licence does not relicense the underlying music compositions or recordings. Credits and source material for the included arrangements:

- "Star Trek: The Next Generation": main-title music by Jerry Goldsmith, incorporating Alexander Courage's Star Trek fanfare. MIDI source: [S16.MID](https://www.vic-fontaine.com/countm.php?group=10&nbr=3), listed on [Vic Fontaine's TNG page](https://www.vic-fontaine.com/tng.php). The two controller tracks retain the highest brass voice and low string voice, with the silent introduction removed. Duration: 70.16 seconds.
- "Star Trek opening fanfare": music by Alexander Courage. MIDI source: [STARTREK.MID](https://www.vic-fontaine.com/countm.php?group=2&nbr=8), listed on [Vic Fontaine's original-series page](https://www.vic-fontaine.com/tos.php). The controller arrangement contains the first complete eight-note brass call, 6.08 seconds.
- "Knight Rider" and "robot car for justice": two-voice arrangements of the Knight Rider theme by Stu Phillips and Glen A. Larson. The latter is the compact riff-and-fanfare arrangement.
- "TNG Warp 2": [TrekCore's TNG Warp 2 clean sound](https://www.trekcore.com/audio/warp/tng_warp2_clean.mp3), converted to signed 8-bit mono at 16000 Hz, without normalization. The resident stream has 44160 samples and SHA-256 `497f2f7bd55a11e8232a83a62792c7d7868650487266a35fa1b0d405ab3139b3`.
- "Axel F": music by Harold Faltermeyer. This entry uses the native package already installed in the controller; its MIDI/native library is not included here.

MIDI tracks are embedded in the service binary and copied into its persistent song library. The warp stream is supplied separately in `controller-samples/` and in the release's controller-sample archive. It must already be resident in a compatible VESC app; installing the MDB service does not upload it or flash the controller.

The original multitrack MIDI files and MP3 are not included. Controller tracks are type-0, single-track MIDI files under the 2048-byte upload limit.
