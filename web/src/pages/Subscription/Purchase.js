import React, { useEffect, useState } from 'react';
import { API, showError, showSuccess } from '../../helpers';

import {
  Package,
  DollarSign,
  Calendar,
  CheckCircle,
  Star,
  Zap,
  Shield,
  CreditCard
} from 'lucide-react';
import {
  Button,
  Card,
  Col,
  Row,
  Space,
  Typography,
  Modal,
  Form,
  Select,
  Spin,
  Tag,
  Divider,
  List,
  Badge
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const { Text, Title } = Typography;

const SubscriptionPurchase = () => {
  const { t } = useTranslation();
  const [plans, setPlans] = useState([]);
  const [loading, setLoading] = useState(true);
  const [purchasing, setPurchasing] = useState(false);
  const [showPurchaseModal, setShowPurchaseModal] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState(null);
  const [userQuotas, setUserQuotas] = useState({});

  const loadPlans = async () => {
    try {
      const res = await API.get('/api/subscription/plans?status=1');
      const { success, message, data } = res.data;
      if (success) {
        setPlans(data || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const loadUserQuotas = async () => {
    try {
      const res = await API.get('/api/subscription/quotas');
      const { success, data } = res.data;
      if (success) {
        setUserQuotas(data.quotas || {});
      }
    } catch (error) {
      // 用户可能没有订阅，忽略错误
    }
  };

  const loadData = async () => {
    setLoading(true);
    await Promise.all([loadPlans(), loadUserQuotas()]);
    setLoading(false);
  };

  useEffect(() => {
    loadData();
  }, []);

  const handlePurchase = async (values) => {
    setPurchasing(true);
    try {
      const res = await API.post('/api/subscription/purchase', {
        plan_id: selectedPlan.id,
        payment_method: values.payment_method
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('购买成功！'));
        setShowPurchaseModal(false);
        loadUserQuotas(); // 重新加载用户配额
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setPurchasing(false);
    }
  };

  const renderPlanCard = (plan) => {
    const quotas = JSON.parse(plan.model_quotas || '{}');
    const quotaList = Object.entries(quotas);
    
    return (
      <Card
        key={plan.id}
        style={{ height: '100%' }}
        bodyStyle={{ padding: 24 }}
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        title={
          <Space>
            <Package size={20} />
            <Text strong size="large">{plan.name}</Text>
            <Badge count="热门" type="danger" style={{ marginLeft: 8 }} />
          </Space>
        }
      >
        <Space vertical style={{ width: '100%' }} spacing={16}>
          {/* 价格 */}
          <div style={{ textAlign: 'center' }}>
            <Space align="baseline">
              <Text size="large" type="secondary">¥</Text>
              <Title heading={2} style={{ margin: 0, color: '#1890ff' }}>
                {plan.price}
              </Title>
              <Text type="secondary">/ {plan.duration}天</Text>
            </Space>
          </div>

          {/* 描述 */}
          {plan.description && (
            <Text type="secondary" style={{ textAlign: 'center' }}>
              {plan.description}
            </Text>
          )}

          <Divider margin={12} />

          {/* 配额列表 */}
          <div>
            <Text strong style={{ marginBottom: 8, display: 'block' }}>
              <Zap size={16} style={{ marginRight: 4 }} />
              {t('包含配额')}
            </Text>
            <List
              size="small"
              dataSource={quotaList}
              renderItem={([model, quota]) => (
                <List.Item style={{ padding: '8px 0' }}>
                  <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                    <Text>{model}</Text>
                    <Tag color="blue" size="small">{quota}次</Tag>
                  </Space>
                </List.Item>
              )}
            />
          </div>

          {/* 特性 */}
          <div>
            <Text strong style={{ marginBottom: 8, display: 'block' }}>
              <Star size={16} style={{ marginRight: 4 }} />
              {t('套餐特性')}
            </Text>
            <Space vertical spacing={4}>
              <Space>
                <CheckCircle size={14} color="#52c41a" />
                <Text size="small">优先使用订阅配额</Text>
              </Space>
              <Space>
                <CheckCircle size={14} color="#52c41a" />
                <Text size="small">配额用完自动切换按量计费</Text>
              </Space>
              <Space>
                <CheckCircle size={14} color="#52c41a" />
                <Text size="small">实时配额监控</Text>
              </Space>
              <Space>
                <Shield size={14} color="#52c41a" />
                <Text size="small">7x24小时技术支持</Text>
              </Space>
            </Space>
          </div>

          {/* 购买按钮 */}
          <Button
            type="primary"
            size="large"
            block
            icon={<CreditCard size={16} />}
            onClick={() => {
              setSelectedPlan(plan);
              setShowPurchaseModal(true);
            }}
          >
            {t('立即购买')}
          </Button>
        </Space>
      </Card>
    );
  };

  if (loading) {
    return (
      <Card style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
        <Text style={{ display: 'block', marginTop: 16 }}>
          {t('加载套餐信息中...')}
        </Text>
      </Card>
    );
  }

  return (
    <Space vertical style={{ width: '100%' }} spacing={24}>
      {/* 页面标题 */}
      <Card>
        <Space vertical style={{ width: '100%' }} spacing={16}>
          <div style={{ textAlign: 'center' }}>
            <Title heading={3}>
              <Package size={24} style={{ marginRight: 8 }} />
              {t('订阅套餐')}
            </Title>
            <Text type="secondary">
              {t('选择适合您的订阅套餐，享受更优惠的AI服务')}
            </Text>
          </div>

          {/* 当前配额状态 */}
          {Object.keys(userQuotas).length > 0 && (
            <Card size="small" style={{ background: '#f8f9fa' }}>
              <Text strong style={{ marginBottom: 8, display: 'block' }}>
                {t('当前订阅配额')}
              </Text>
              <Row gutter={16}>
                {Object.entries(userQuotas).map(([model, quota]) => (
                  <Col span={6} key={model}>
                    <Space vertical spacing={4}>
                      <Text size="small" type="secondary">{model}</Text>
                      <Text strong>{quota.remaining}/{quota.total}</Text>
                    </Space>
                  </Col>
                ))}
              </Row>
            </Card>
          )}
        </Space>
      </Card>

      {/* 套餐列表 */}
      <Row gutter={[24, 24]}>
        {plans.map(plan => (
          <Col span={8} key={plan.id}>
            {renderPlanCard(plan)}
          </Col>
        ))}
      </Row>

      {/* 购买确认弹窗 */}
      <Modal
        title={t('确认购买')}
        visible={showPurchaseModal}
        onCancel={() => setShowPurchaseModal(false)}
        footer={null}
        width={500}
      >
        {selectedPlan && (
          <Form onSubmit={handlePurchase} labelPosition="left" labelWidth={80}>
            <Space vertical style={{ width: '100%' }} spacing={16}>
              <Card size="small" style={{ background: '#f8f9fa' }}>
                <Space vertical spacing={8}>
                  <Text strong>{selectedPlan.name}</Text>
                  <Text type="secondary">{selectedPlan.description}</Text>
                  <Space>
                    <DollarSign size={16} />
                    <Text strong>¥{selectedPlan.price}</Text>
                    <Calendar size={16} />
                    <Text>{selectedPlan.duration}天</Text>
                  </Space>
                </Space>
              </Card>

              <Form.Select
                field="payment_method"
                label={t('支付方式')}
                placeholder={t('请选择支付方式')}
                rules={[{ required: true, message: t('请选择支付方式') }]}
                optionList={[
                  { label: t('余额支付'), value: 'balance' },
                  { label: t('微信支付'), value: 'wechat' },
                  { label: t('支付宝'), value: 'alipay' },
                ]}
              />

              <div style={{ textAlign: 'right' }}>
                <Space>
                  <Button onClick={() => setShowPurchaseModal(false)}>
                    {t('取消')}
                  </Button>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={purchasing}
                  >
                    {t('确认购买')}
                  </Button>
                </Space>
              </div>
            </Space>
          </Form>
        )}
      </Modal>
    </Space>
  );
};

export default SubscriptionPurchase;
