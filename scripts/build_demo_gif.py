from pathlib import Path

from PIL import Image


def main() -> None:
    root = Path(__file__).resolve().parent.parent
    frames_dir = root / "docs" / "media" / "frames"
    output = root / "docs" / "media" / "demo.gif"

    frame_paths = sorted(frames_dir.glob("frame-*.png"))
    if not frame_paths:
        raise SystemExit("no frames found in docs/media/frames")

    frames = [Image.open(p).convert("P", palette=Image.ADAPTIVE) for p in frame_paths]
    frames[0].save(
        output,
        save_all=True,
        append_images=frames[1:],
        duration=900,
        loop=0,
        optimize=True,
    )

    print(output)


if __name__ == "__main__":
    main()
