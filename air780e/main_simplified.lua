PROJECT = "sms"
VERSION = "1.0.0"

log.info("main", PROJECT, VERSION)

sys = require("sys")
require "sysplus"

if wdt then
  wdt.init(9000)
  sys.timerLoopStart(wdt.feed, 3000)
end

----------------------------------------------------------------
-- Protocol definitions
PROTOCOL_VERSION = 1
FLAG_HEARTBEAT = 1
FLAG_HEARTBEAT_REQUEST = 2
ENCRYPT_NONE = 0

TAG_SMS_RECEIVED = 1
TAG_SMS_SEND = 2
TAG_SMS_ACK = 3

----------------------------------------------------------------
-- Message handling

function msg_handler(data)
  local msg = json.decode(data)
  if msg == nil then
    log.info("msg_handler: invalid json", data)
    return
  end

  local md5 = crypto.md5(msg.data)
  if md5 ~= msg.md5 then
    log.info("msg_handler: md5 mismatch", md5, msg.md5)
    return
  end

  if msg.tag == TAG_SMS_SEND then
    local vsms = json.decode(msg.data)
    if vsms == nil then
      log.info("msg_handler: invalid sms data", msg.data)
      return
    end
    log.info("Sending SMS to", vsms.phone)
    local res = sms.send(vsms.phone, vsms.msg)
    if res then
      msg_send(TAG_SMS_ACK, json.encode({key=md5}))
    end
    return
  end
end

function msg_send(tag, data)
  local pkg = string.char(0xff, 0x07, 0x55, 0x00)
  pkg = pkg .. string.char(PROTOCOL_VERSION)
  local msg = json.encode({tag=tag, md5=crypto.md5(data), data=data})
  local headFooter = string.char(0, ENCRYPT_NONE, 0) .. Int32ToBuf(#msg) .. Int32ToBuf(crypto.crc32(msg))
  pkg = pkg .. Int32ToBuf(crypto.crc32(headFooter)) .. headFooter .. msg
  uart.write(UART_ID, pkg)
end

function heartbeat_response()
  log.info("Sending heartbeat response")
  local pkg = string.char(0xff, 0x07, 0x55, 0x00)
  pkg = pkg .. string.char(PROTOCOL_VERSION)
  local headFooter = string.char(FLAG_HEARTBEAT, ENCRYPT_NONE, 0) .. Int32ToBuf(0) .. Int32ToBuf(crypto.crc32(""))
  pkg = pkg .. Int32ToBuf(crypto.crc32(headFooter)) .. headFooter
  uart.write(UART_ID, pkg)
end

----------------------------------------------------------------
-- UART Configuration

UART_ID = 1

local uart_ok = uart.setup(
    UART_ID,
    115200,
    8,
    1,
    uart.NONE,
    uart.LSB,
    4112
)

PACKAGE_MAX_SIZE = 4096
PACKAGE_HEAD_SIZE = 20

HEAD_OFFSET_PREFIX = 0 + 1
HEAD_OFFSET_VERSION = 4 + 1
HEAD_OFFSET_CRC32 = 5 + 1
HEAD_OFFSET_FLAG = 9 + 1
HEAD_OFFSET_ENCRYPT = 10 + 1
HEAD_OFFSET_VALUE = 11 + 1
HEAD_OFFSET_DATA_SIZE = 12 + 1
HEAD_OFFSET_DATA_CRC32 = 16 + 1
HEAD_OFFSET_DATA = 20 + 1

r_buf = ""

uart.on(UART_ID, "receive", function(id, len)
  repeat
    local s = uart.read(id, len)
    r_buf = s
    if #r_buf >= PACKAGE_HEAD_SIZE then
      if string.byte(r_buf, HEAD_OFFSET_PREFIX) ~= 0xff then
        log.info("uart: invalid prefix")
        break
      end
      if string.byte(r_buf, HEAD_OFFSET_PREFIX + 1) ~= 0x07 then
        break
      end
      if string.byte(r_buf, HEAD_OFFSET_PREFIX + 2) ~= 0x55 then
        break
      end
      if string.byte(r_buf, HEAD_OFFSET_PREFIX + 3) ~= 0x00 then
        break
      end

      if string.byte(r_buf, HEAD_OFFSET_VERSION) ~= PROTOCOL_VERSION then
        log.info("uart: version mismatch")
        break
      end

      local crc32 = bufToUInt32(string.sub(r_buf, HEAD_OFFSET_CRC32, HEAD_OFFSET_CRC32 + 3))
      local crc32_got = crypto.crc32(string.sub(r_buf, HEAD_OFFSET_FLAG, HEAD_OFFSET_DATA-1))
      if crc32 ~= crc32_got then
        log.info("uart: crc32 mismatch")
        break
      end

      if (string.byte(r_buf, HEAD_OFFSET_FLAG) & FLAG_HEARTBEAT_REQUEST) ~= 0 then
        log.info("uart: heartbeat request received")
        heartbeat_response()
      end

      if string.byte(r_buf, HEAD_OFFSET_ENCRYPT) ~= ENCRYPT_NONE then
        log.info("uart: unsupported encryption")
        break
      end

      local data_size = bufToUInt32(string.sub(r_buf, HEAD_OFFSET_DATA_SIZE, HEAD_OFFSET_DATA_SIZE + 3))
      local data_crc32 = bufToUInt32(string.sub(r_buf, HEAD_OFFSET_DATA_CRC32, HEAD_OFFSET_DATA_CRC32 + 3))
      local data = string.sub(r_buf, HEAD_OFFSET_DATA, HEAD_OFFSET_DATA+data_size)
      local data_crc32_got = crypto.crc32(data)
      if data_crc32 ~= data_crc32_got then
        log.info("uart: data crc32 mismatch")
        break
      end
      msg_handler(data)
    end
    if #s == len then
      break
    end
  until s == ""
end)

----------------------------------------------------------------
-- SMS Handler (simplified - only forwards to Raspberry Pi)

function sms_handler(num, txt, meta)
  log.info("SMS received from", num, ":", txt)

  local body = json.encode({
    phone=num,
    msg=txt,
    time=string.format("20%02d-%02d-%02d %02d:%02d:%02d",
      meta.year, meta.mon, meta.day,
      meta.hour, meta.min, meta.sec)
  })

  msg_send(TAG_SMS_RECEIVED, body)
end

sms.setNewSmsCb(sms_handler)

----------------------------------------------------------------
-- Utility Functions

function bufToUInt32(buf)
  return (string.byte(buf, 1) << 24) | (string.byte(buf, 2) << 16) | (string.byte(buf, 3) << 8) | (string.byte(buf, 4))
end

function Int32ToBuf(n)
  return (string.format("%c%c%c%c",n >> 24,n >> 16,n >> 8,n))
end

----------------------------------------------------------------
-- Main Loop

sys.taskInit(function()
  sys.wait(5000)
  log.info("SMS service started")
end)

sys.run()