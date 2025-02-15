PROJECT = "sms"
VERSION = "0.0.2"

log.info("main", PROJECT, VERSION)

sys = require("sys")
require "sysplus" -- http库需要这个sysplus

if wdt then
  --添加硬狗防止程序卡死，在支持的设备上启用这个功能
  wdt.init(9000)--初始化watchdog设置为9s
  sys.timerLoopStart(wdt.feed, 3000)--3s喂一次狗
end

----------------------------------------------------------------
-- 通用消息处理
--
-- 与上位机数据通信格式，转化为base64格式发送
-- {tag:int, md5:string, data=string}
--   tag: int类型数据，表示不同的数据格式
--   md5: string类型数据，表示数据的md5校验值
--   data: string类型数据，数据base64化后的字符串

PROTOCOL_VERSION = 1
FLAG_HEARTBEAT = 1
FLAG_HEARTBEAT_REQUEST = 2
ENCRYPT_NONE = 0

TAG_SMS_RECEIVED = 1
TAG_SMS_SEND = 2
TAG_SMS_ACK = 3

-- 处理接收到的通用消息
--   data:string 通用消息
function msg_handler(data)
  local msg = json.decode(data) -- 解析json
  if msg == nil then
    log.info(" body == nil", data)
    return
  end
  -- 校验数据md5
  local md5 = crypto.md5(msg.data)
  if md5 ~= msg.md5 then
    log.info(" md5 ~= msg.md5", md5, msg.md5)
    return
  end
  -- 消息处理
  if msg.tag == TAG_SMS_SEND then
    -- 解析sms
    local vsms = json.decode(msg.data)
    if vsms == nil then
      log.info(" sms == nil", msg.data)
      return
    end
    log.info("send sms: ", vsms.phone, vsms.msg)
    local res = sms.send(vsms.phone, vsms.msg)
    if res then
      msg_send(TAG_SMS_ACK, json.encode({key=md5}))
    end
    return
  end
  log.info("sms_handler", data)
end

-- 发送通用消息
--   tag:int 消息类型
--   data:string 消息数据

function msg_send(tag, data)
  --log.info("msg_send", data)
  local pkg = string.char(0xff, 0x07, 0x55, 0x00)
  pkg = pkg .. string.char(PROTOCOL_VERSION)
  local msg = json.encode({tag=tag, md5=crypto.md5(data), data=data})
  local headFooter = string.char(0, ENCRYPT_NONE, 0) .. Int32ToBuf(#msg) .. Int32ToBuf(crypto.crc32(msg))
  pkg = pkg .. Int32ToBuf(crypto.crc32(headFooter)) .. headFooter .. msg
  --log.info("msg_send", #pkg, pkg:toHex())
  uart.write(UART_ID, pkg)
end

function heartbeat_response()
  log.info("[heartbeat_send] response send")
  local pkg = string.char(0xff, 0x07, 0x55, 0x00)
  pkg = pkg .. string.char(PROTOCOL_VERSION)
  local headFooter = string.char(FLAG_HEARTBEAT, ENCRYPT_NONE, 0) .. Int32ToBuf(0) .. Int32ToBuf(crypto.crc32(""))
  pkg = pkg .. Int32ToBuf(crypto.crc32(headFooter)) .. headFooter
  --log.info("heartbeat_send", #pkg, pkg:toHex())
  uart.write(UART_ID, pkg)
end

----------------------------------------------------------------
-- UART

UART_ID = 1

-- 初始化UART
local uart_ok = uart.setup(
    UART_ID, --串口id
    115200, --波特率
    8, --数据位
    1, --停止位
    uart.NONE, --校验位
    uart.LSB, --大小端
    4112 --缓冲区大小
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

ENCRYPT_NONE = 0

r_buf = ""

-- 设置UART回调函数
uart.on(UART_ID, "receive", function(id, len)
  repeat
    local s = uart.read(id, len)
    r_buf = s
    if #r_buf >= PACKAGE_HEAD_SIZE then
      -- prefix
      if string.byte(r_buf, HEAD_OFFSET_PREFIX) ~= 0xff then
        log.info(" uart: error on 0xff")
        break
      end
      if string.byte(r_buf, HEAD_OFFSET_PREFIX + 1) ~= 0x07 then
        log.info(" uart: error on 0x07")
        break
      end
      if string.byte(r_buf, HEAD_OFFSET_PREFIX + 2) ~= 0x55 then
        log.info(" uart: error on 0x55")
        break
      end
      if string.byte(r_buf, HEAD_OFFSET_PREFIX + 3) ~= 0x00 then
        log.info(" uart: error on 0x00")
        break
      end
      -- version
      if string.byte(r_buf, HEAD_OFFSET_VERSION) ~= PROTOCOL_VERSION then
        log.info(" uart: error on prot version")
        break
      end
      -- crc32
      local crc32 = bufToUInt32(string.sub(r_buf, HEAD_OFFSET_CRC32, HEAD_OFFSET_CRC32 + 3))
      local crc32_got = crypto.crc32(string.sub(r_buf, HEAD_OFFSET_FLAG, HEAD_OFFSET_DATA-1))
      if crc32 ~= crc32_got then
        log.info(string.format("%x%x%x%x", string.byte(r_buf, HEAD_OFFSET_CRC32), string.byte(r_buf, HEAD_OFFSET_CRC32+1), string.byte(r_buf, HEAD_OFFSET_CRC32+2), string.byte(r_buf, HEAD_OFFSET_CRC32+3)))
        log.info(" uart: error on crc32, need ", string.format("%x", crc32), " got ", string.format("%x", crc32_got))
        break
      end
      -- flag
      if (string.byte(r_buf, HEAD_OFFSET_FLAG) & FLAG_HEARTBEAT) ~= 0 then
        log.info(" uart: heartbeat received")
        heartbeat_received = true
      end
      if (string.byte(r_buf, HEAD_OFFSET_FLAG) & FLAG_HEARTBEAT_REQUEST) ~= 0 then
        log.info(" uart: heartbeat request received")
        heartbeat_received = true
        heartbeat_response()
      end
      -- encrypt method
      if string.byte(r_buf, HEAD_OFFSET_ENCRYPT) ~= ENCRYPT_NONE then
        log.info(" uart: error on encrypt method")
        break
      end
      -- value
      -- data size
      local data_size = bufToUInt32(string.sub(r_buf, HEAD_OFFSET_DATA_SIZE, HEAD_OFFSET_DATA_SIZE + 3))
      local data_crc32 = bufToUInt32(string.sub(r_buf, HEAD_OFFSET_DATA_CRC32, HEAD_OFFSET_DATA_CRC32 + 3))
      local data = string.sub(r_buf, HEAD_OFFSET_DATA, HEAD_OFFSET_DATA+data_size)
      local data_crc32_got = crypto.crc32(data)
      if data_crc32 ~= data_crc32_got then
        log.info(" uart: error on data crc32")
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
-- SMS

-- 接受短信回调函数
--   1. 通过uart发送到ha，由ha进一步处理
--   num 手机号码
--   txt 文本内容
function sms_handler(num, txt, meta)
  if txt == "help" then
    sms.send("+8612345678900", "[SMS][HELP]\nreboot - Reboot SMS\nstatus - SMS Status\ncstatus - SMS Current Status")
    return
  end
  if txt == "reboot" then
    sms.send("+8612345678900", "[SMS][Reboot]")
    sys.wait(3000)
    pm.reboot()
    return
  end
  if txt == "status" then
    send_status()
    return
  end
  if txt == "cstatus" then
    send_status()
    return
  end
  local body = json.encode({phone=num, msg=txt, time=string.format("20%02d-%02d-%02d %02d:%02d:%02d", meta.year, meta.mon, meta.day, meta.hour, meta.min, meta.sec)}) -- 短信数据json
  msg_send(TAG_SMS_RECEIVED, body)
end
-- 设置短信回调函数
sms.setNewSmsCb(sms_handler)

----------------------------------------------------------------
-- HEARTBEAT

function send_status()
  status="[SMS][HA]"
  if power_last_state == 0 then
    status=status .. " down"
  elseif power_last_state == 1 then
    status=status .. " up"
  else
    status=status .. " nil"
  end
  if ha_last_state then
    status=status .. " connected"
  else
    status=status .. " disconnected"
  end
  sms.send("+8612345678900", status)
end

function send_cstatus()
  status="[SMS][HA]"
  power_state = gpio.get(GPIO_HA_POWER_PIN)
  if power_state == 0 then
    status=status .. " down"
  elseif power_state == 1 then
    status=status .. " up"
  else
    status=status .. " nil"
  end
  if heartbeat_received then
    status=status .. " connected"
  else
    status=status .. " disconnected"
  end
  sms.send("+8612345678900", status)
end

GPIO_HA_POWER_PIN = 1
gpio.setup(GPIO_HA_POWER_PIN, nil, gpio.PULLUP)
-- power last state 电源上一次的状态
-- 0 power down
-- 1 power up
power_last_state = 1
power_last_state_times = -1

function check_power()
  power_state = gpio.get(GPIO_HA_POWER_PIN)
  log.info("[check_connection] GPIO ", string.format("%d", power_state))
  if power_state ~= power_last_state then
    if power_last_state_times == 0 then
      -- 状态仍是变化后的，等待次数为0，发出提醒
    	send_healthy_power()
    	-- 修改上次的状态，更改为当前状态，不再发出此状态的提醒
    	power_last_state = power_state
    	power_last_state_times = -1
    elseif power_last_state_times == -1 then
      -- 状态刚发生变化，初始化次数等待发送信息
      power_last_state_times = 3
    else
      -- 状态仍是变化后的，减少等待次数
      power_last_state_times = power_last_state_times - 1
    end
  else
    -- 结果与上次相同，忽略（如果中间掺杂其他状态，忽略
    power_last_state_times = -1
  end
end

function send_healthy_power()
	if power_state == 0 then
    sms.send("+8612345678900", "[SMS][HA] power down")
  else
    sms.send("+8612345678900", "[SMS][HA] power up")
  end
end

heartbeat_received = false
-- ha last state HA上一次的状态
--  true  connected 与ha建立连接
--  false disconnected 与ha断开连接
ha_last_state = true
ha_last_state_times = -1

function check_connection()
  ha_state = false
  if heartbeat_received then
    log.info("[check_connection] heartbeat received")
    heartbeat_received = false
    ha_state = true
  else
    log.info("[check_connection] heartbeat missing")
    ha_state = false
  end
  if ha_state ~= ha_last_state then
    if ha_last_state_times == 0 then
      -- 状态仍是变化后的，等待次数为0，发出提醒
    	send_healthy_ha()
    	-- 修改上次的状态，更改为当前状态，不再发出此状态的提醒
    	ha_last_state = ha_state
    	ha_last_state_times = -1
    elseif ha_last_state_times == -1 then
      -- 状态刚发生变化，初始化次数等待发送信息
      power_last_state_times = 3
    else
      -- 状态仍是变化后的，减少等待次数
      power_last_state_times = power_last_state_times - 1
    end
  else
    -- 结果与上次相同，忽略（如果中间掺杂其他状态，忽略
    ha_last_state_times = -1
  end
end

function send_healthy_ha()
	if ha_last_state then
    sms.send("+8612345678900", "[SMS][HA] connected")
  else
    sms.send("+8612345678900", "[SMS][HA] disconnected")
  end
end

-- 每10s检查电源是否正常
sys.timerLoopStart(check_power, 10000)
-- 每60s检查连接是否正常
sys.timerLoopStart(check_connection, 60000)
-- 每8s发送一次心跳信号
--sys.timerLoopStart(heartbeat_send, 8000)

----------------------------------------------------------------
-- UTIL

function bufToUInt32(buf)
  return (string.byte(buf, 1) << 24) | (string.byte(buf, 2) << 16) | (string.byte(buf, 3) << 8) | (string.byte(buf, 4))
end

function Int32ToBuf(n)
  return (string.format("%c%c%c%c",n >> 24,n >> 16,n >> 8,n))
end

----------------------------------------------------------------
-- MAIN

sys.taskInit(function()
  sys.wait(60000)
  send_status()
end)

sys.run()
