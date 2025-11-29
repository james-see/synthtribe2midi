# Homebrew formula for synthtribe2midi
# This file is auto-updated by GoReleaser
class Synthtribe2midi < Formula
  desc "Convert between MIDI and Behringer SynthTribe formats"
  homepage "https://james-see.github.io/synthtribe2midi"
  license "MIT"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/james-see/synthtribe2midi/releases/download/v#{version}/synthtribe2midi_Darwin_arm64.tar.gz"
      # sha256 will be auto-populated by GoReleaser
    else
      url "https://github.com/james-see/synthtribe2midi/releases/download/v#{version}/synthtribe2midi_Darwin_x86_64.tar.gz"
      # sha256 will be auto-populated by GoReleaser
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/james-see/synthtribe2midi/releases/download/v#{version}/synthtribe2midi_Linux_arm64.tar.gz"
      # sha256 will be auto-populated by GoReleaser
    else
      url "https://github.com/james-see/synthtribe2midi/releases/download/v#{version}/synthtribe2midi_Linux_x86_64.tar.gz"
      # sha256 will be auto-populated by GoReleaser
    end
  end

  def install
    bin.install "synthtribe2midi"
    bin.install "synthtribe2midi-server"
  end

  test do
    system "#{bin}/synthtribe2midi", "--version"
  end
end

