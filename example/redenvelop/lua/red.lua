-- KEY[1]: 用户防重领取记录
local userHashKey = KEYS[1];
-- KEY[2]: 运营预分配红包列表
local redPacketOperatingKey = KEYS[2];
-- KEY[3]: 用户红包领取记录
local userAmountKey = KEYS[3];
-- KEY[4]: 用户编号
local userId = KEYS[4];
local result = {};
-- 判断用户是否领取过
if redis.call('hexists', userHashKey, userId) == 1 then
    result['msg'] = '已经领取了';
    result['code'] = 1;
    return cjson.encode(result);
else
    -- 从预分配红包中获取红包数据
    local redPacket = redis.call('rpop', redPacketOperatingKey);
    if redPacket then
        local data = cjson.decode(redPacket);
        -- 加入用户ID信息
        data['userId'] = userId;
        -- 把用户编号放到去重hash value 设置位红包编号
        redis.call('hset', userHashKey, userId, data['redPacketId']);
        -- 用户和红包放到已消费队列里
        redis.call('lpush', userAmountKey, cjson.encode(data));
        -- 组装返回结果值
        result['redPacketId'] = data['redPacketId'];
        result['amount'] = data['amount'];
        result['code'] = 0;
        return cjson.encode(result)
    else
        -- 抢红包失败
        result['msg'] = '抢红包失败';
        result['code'] = 1;
        return cjson.encode(result);
    end
end