/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package buffer

import (
	"io"

	"github.com/alipay/sofamosn/pkg/types"
)

type WatermarkBuffer struct {
	IoBuffer

	highWatermark            uint32
	lowWatermark             uint32
	aboveHighWatermarkCalled bool
	watermarkListener        types.BufferWatermarkListener
}

func (b *WatermarkBuffer) Read(p []byte) (n int, err error) {
	n, err = b.IoBuffer.Read(p)
	b.checkLowWatermark()

	return
}

func (b *WatermarkBuffer) ReadOnce(r io.Reader) (n int64, err error) {
	n, err = b.IoBuffer.ReadOnce(r)
	b.checkHighWatermark()

	return
}

func (b *WatermarkBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = b.IoBuffer.ReadFrom(r)
	b.checkHighWatermark()

	return
}

func (b *WatermarkBuffer) Write(p []byte) (n int, err error) {
	n, err = b.IoBuffer.Write(p)
	b.checkHighWatermark()

	return
}

func (b *WatermarkBuffer) WriteTo(w io.Writer) (n int64, err error) {
	n, err = b.IoBuffer.WriteTo(w)
	b.checkLowWatermark()

	return
}

func (b *WatermarkBuffer) Append(data []byte) (err error) {
	err = b.IoBuffer.Append(data)
	b.checkHighWatermark()

	return
}

func (b *WatermarkBuffer) AppendByte(data byte) (err error) {
	err = b.IoBuffer.AppendByte(data)
	b.checkHighWatermark()

	return
}

func (b *WatermarkBuffer) Cut(offset int) (ioBuffer types.IoBuffer) {
	ioBuffer = b.IoBuffer.Cut(offset)
	b.checkLowWatermark()

	return
}

func (b *WatermarkBuffer) Drain(offset int) {
	b.IoBuffer.Drain(offset)
	b.checkLowWatermark()

	return
}

func (b *WatermarkBuffer) Reset() {
	b.IoBuffer.Reset()
	b.checkLowWatermark()
}

func (b *WatermarkBuffer) checkLowWatermark() {
	if !b.aboveHighWatermarkCalled ||
		(b.highWatermark > 0 && uint32(b.Len()) >= b.lowWatermark) {
		return
	}

	b.aboveHighWatermarkCalled = false
	b.watermarkListener.OnLowWatermark()
}

func (b *WatermarkBuffer) checkHighWatermark() {
	if b.aboveHighWatermarkCalled || b.highWatermark == 0 ||
		uint32(b.Len()) <= b.highWatermark {
		return
	}

	b.aboveHighWatermarkCalled = true
	b.watermarkListener.OnHighWatermark()
}

func (b *WatermarkBuffer) SetWaterMark(watermark uint32) {
	b.SetWaterMarks(watermark/2, watermark)
}

func (b *WatermarkBuffer) SetWaterMarks(lowWatermark uint32, highWatermark uint32) {
	b.lowWatermark = lowWatermark
	b.highWatermark = highWatermark
}

func NewWatermarkBuffer(capacity int, listener types.BufferWatermarkListener) types.IoBuffer {
	return &WatermarkBuffer{
		IoBuffer: IoBuffer{
			buf:     make([]byte, 0, capacity),
			offMark: ResetOffMark,
		},
		watermarkListener: listener,
	}
}

func NewWatermarkBufferString(s string, listener types.BufferWatermarkListener) types.IoBuffer {
	return &WatermarkBuffer{
		IoBuffer: IoBuffer{
			buf:     []byte(s),
			offMark: ResetOffMark,
		},
		watermarkListener: listener,
	}
}

func NewWatermarkBufferBytes(bytes []byte, listener types.BufferWatermarkListener) types.IoBuffer {
	return &WatermarkBuffer{
		IoBuffer: IoBuffer{
			buf:     []byte(bytes),
			offMark: ResetOffMark,
		},
		watermarkListener: listener,
	}
}
