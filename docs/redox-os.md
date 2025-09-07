Here’s the easiest way to boot Redox OS on an M2 Mac with QEMU. You have two paths:

⸻

Option A — Emulate the x86_64 image (works today, slower)
1.	Install tools

brew install qemu zstd

	2.	Download & decompress a prebuilt Redox x86_64 image (demo is fine)

# grab a demo harddrive image (.img.zst) from the releases or daily builds
# then:
zstd -d ~/Downloads/redox_demo_x86_64_*_harddrive.img.zst

Docs & links are here.  ￼
3.	Run it (x86_64 is TCG-emulated on Apple Silicon; no HVF accel for x86)

SDL_VIDEO_X11_DGAMOUSE=0 qemu-system-x86_64 \
-d cpu_reset,guest_errors -smp 4 -m 2048 \
-chardev stdio,id=debug,signal=off,mux=on,"" \
-serial chardev:debug -mon chardev=debug \
-machine q35 -device ich9-intel-hda -device hda-duplex \
-netdev user,id=net0 -device e1000,netdev=net0 \
-device nec-usb-xhci,id=xhci \
-cpu max \
-drive file="$(echo ~/Downloads/redox_demo_x86_64_*_harddrive.img)",format=raw

(Command adapted from the official docs; on Apple Silicon this runs under TCG, not HVF.)  ￼ ￼

Login tips: when it boots, choose a resolution; then log in as user (no password) or root (password).  ￼

⸻

Option B — Run a native AArch64 build (faster with HVF)

Redox ARM64 support exists but is more limited than x86_64. There are daily aarch64 images you can try.
1.	Install tools & locate UEFI firmware

brew install qemu zstd
# firmware path from Homebrew QEMU:
ls /opt/homebrew/share/qemu/edk2-aarch64-code.fd

(That file is the AArch64 UEFI firmware QEMU uses.)  ￼ ￼
2.	Download & decompress an aarch64 image

# from the aarch64 daily images index (demo/desktop/server/minimal)
# e.g. redox_demo_aarch64_*_harddrive.img.zst
zstd -d ~/Downloads/redox_demo_aarch64_*_harddrive.img.zst

(Daily AArch64 images are published here.)  ￼
3.	Run it with ARM virtualization (HVF)

qemu-system-aarch64 \
-M virt -accel hvf -cpu host -smp 4 -m 2048 \
-bios /opt/homebrew/share/qemu/edk2-aarch64-code.fd \
-device virtio-gpu-pci -device qemu-xhci -device usb-kbd -device usb-mouse \
-netdev user,id=net0 -device virtio-net-pci,netdev=net0 \
-drive file="$(echo ~/Downloads/redox_demo_aarch64_*_harddrive.img)",format=raw

(Use the virt machine, the edk2 AArch64 firmware, and HVF acceleration on Apple Silicon.)  ￼ ￼

Note: Redox’s ARM64 has limited support; you may prefer the x86_64 image if you hit issues.  ￼

⸻

Which should you pick?
•	Want it to “just work”? Use Option A (x86_64) — slower but most tested.  ￼
•	Want speed on M2 and don’t mind rough edges? Try Option B (AArch64) — uses HVF and runs much faster.  ￼ ￼

If you tell me which variant you downloaded (demo/desktop/server & arch), I’ll tailor the exact QEMU command for you.