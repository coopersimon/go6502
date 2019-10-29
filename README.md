# 65C02 processor

A WDC 65C02 processor emulator. It aims to be cycle accurate.

### Usage

It should be constructed with a MemoryBus that implements the interface provided. It can then be driven using the `Step` method. `Step` advances the processor by one instruction and then returns the amount of cycles it took. It also clocks the MemoryBus by that many cycles. In this way you can handle the clock speed from the outside.