import React, { useEffect, useState } from 'react';
import { API, showError, showSuccess, renderNumber } from '../../helpers';

import {
  Users,
  Package,
  DollarSign,
  Calendar,
  CheckCircle,
  XCircle,
  Clock,
  BarChart3,
  Eye
} from 'lucide-react';
import {
  Button,
  Card,
  Divider,
  Empty,
  Form,
  Modal,
  Space,
  Table,
  Tag,
  Typography,
  Tooltip,
  Progress,
  Descriptions
} from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark
} from '@douyinfe/semi-illustrations';
import {
  IconSearch,
  IconRefresh,
  IconEyeOpened
} from '@douyinfe/semi-icons';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

const UserSubscriptionsTable = () => {
  const { t } = useTranslation();
  const [subscriptions, setSubscriptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [total, setTotal] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [showDetail, setShowDetail] = useState(false);
  const [selectedSubscription, setSelectedSubscription] = useState(null);

  const columns = [
    {
      title: t('用户'),
      dataIndex: 'user',
      render: (user) => (
        <Space>
          <Users size={16} />
          <Text strong>{user?.username || '-'}</Text>
        </Space>
      ),
    },
    {
      title: t('套餐'),
      dataIndex: 'subscription_plan',
      render: (plan) => (
        <Space>
          <Package size={16} />
          <Text>{plan?.name || '-'}</Text>
        </Space>
      ),
    },
    {
      title: t('购买价格'),
      dataIndex: 'purchase_price',
      render: (price) => (
        <Space>
          <DollarSign size={14} />
          <Text>¥{price}</Text>
        </Space>
      ),
    },
    {
      title: t('有效期'),
      dataIndex: 'start_time',
      render: (startTime, record) => {
        const start = new Date(startTime * 1000);
        const end = new Date(record.end_time * 1000);
        const now = new Date();
        const isExpired = end < now;
        const remainingDays = Math.max(0, Math.ceil((end - now) / (1000 * 60 * 60 * 24)));
        
        return (
          <Space vertical spacing={4}>
            <Text size="small">
              {start.toLocaleDateString()} - {end.toLocaleDateString()}
            </Text>
            <Tag 
              color={isExpired ? 'red' : remainingDays <= 7 ? 'orange' : 'green'}
              size="small"
            >
              {isExpired ? t('已过期') : `${remainingDays}天后过期`}
            </Tag>
          </Space>
        );
      },
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (status, record) => {
        const now = new Date();
        const endTime = new Date(record.end_time * 1000);
        const isExpired = endTime < now;
        
        if (isExpired && status === 1) {
          return (
            <Tag color='orange' size='large' shape='circle' prefixIcon={<Clock size={14} />}>
              {t('已过期')}
            </Tag>
          );
        }
        
        switch (status) {
          case 1:
            return (
              <Tag color='green' size='large' shape='circle' prefixIcon={<CheckCircle size={14} />}>
                {t('激活')}
              </Tag>
            );
          case 2:
            return (
              <Tag color='red' size='large' shape='circle' prefixIcon={<XCircle size={14} />}>
                {t('过期')}
              </Tag>
            );
          case 3:
            return (
              <Tag color='gray' size='large' shape='circle' prefixIcon={<XCircle size={14} />}>
                {t('取消')}
              </Tag>
            );
          default:
            return (
              <Tag color='gray' size='large' shape='circle'>
                {t('未知')}
              </Tag>
            );
        }
      },
    },
    {
      title: t('配额使用'),
      dataIndex: 'model_quotas',
      render: (quotas, record) => {
        try {
          const remainingQuotas = JSON.parse(quotas || '{}');
          const usedQuotas = JSON.parse(record.used_quotas || '{}');
          const planQuotas = JSON.parse(record.subscription_plan?.model_quotas || '{}');
          
          const models = Object.keys(planQuotas);
          if (models.length === 0) return '-';
          
          let totalUsed = 0;
          let totalQuota = 0;
          
          models.forEach(model => {
            const used = usedQuotas[model] || 0;
            const total = planQuotas[model] || 0;
            totalUsed += used;
            totalQuota += total;
          });
          
          const percentage = totalQuota > 0 ? (totalUsed / totalQuota) * 100 : 0;
          
          return (
            <Tooltip content={
              <div>
                {models.map(model => {
                  const used = usedQuotas[model] || 0;
                  const total = planQuotas[model] || 0;
                  const remaining = remainingQuotas[model] || 0;
                  return (
                    <div key={model}>
                      {model}: {used}/{total} (剩余{remaining})
                    </div>
                  );
                })}
              </div>
            }>
              <Progress
                percent={percentage}
                size="small"
                stroke={percentage > 80 ? '#f5222d' : percentage > 60 ? '#fa8c16' : '#52c41a'}
                showInfo={false}
                style={{ width: 100 }}
              />
            </Tooltip>
          );
        } catch (e) {
          return '-';
        }
      },
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      render: (text) => {
        return new Date(text * 1000).toLocaleString();
      },
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      render: (text, record) => (
        <Button
          theme="borderless"
          icon={<IconEyeOpened />}
          onClick={() => {
            setSelectedSubscription(record);
            setShowDetail(true);
          }}
        >
          {t('查看详情')}
        </Button>
      ),
    },
  ];

  const loadSubscriptions = async (page = 1) => {
    setLoading(true);
    const res = await API.get(`/api/subscription/admin/users?page=${page}&page_size=${ITEMS_PER_PAGE}`);
    const { success, message, data } = res.data;
    if (success) {
      setSubscriptions(data.subscriptions || []);
      setTotal(data.total || 0);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadSubscriptions();
  }, []);

  const handlePageChange = (page) => {
    setActivePage(page);
    loadSubscriptions(page);
  };

  const renderSubscriptionDetail = () => {
    if (!selectedSubscription) return null;
    
    const { subscription_plan, user, model_quotas, used_quotas } = selectedSubscription;
    
    try {
      const remainingQuotas = JSON.parse(model_quotas || '{}');
      const usedQuotasObj = JSON.parse(used_quotas || '{}');
      const planQuotas = JSON.parse(subscription_plan?.model_quotas || '{}');
      
      return (
        <Space vertical style={{ width: '100%' }} spacing={16}>
          <Descriptions
            data={[
              { key: t('用户'), value: user?.username || '-' },
              { key: t('套餐'), value: subscription_plan?.name || '-' },
              { key: t('购买价格'), value: `¥${selectedSubscription.purchase_price}` },
              { key: t('支付方式'), value: selectedSubscription.payment_method || '-' },
              { key: t('开始时间'), value: new Date(selectedSubscription.start_time * 1000).toLocaleString() },
              { key: t('结束时间'), value: new Date(selectedSubscription.end_time * 1000).toLocaleString() },
            ]}
            row
            size="small"
          />
          
          <Card title={t('配额详情')} size="small">
            <Space vertical style={{ width: '100%' }}>
              {Object.entries(planQuotas).map(([model, totalQuota]) => {
                const used = usedQuotasObj[model] || 0;
                const remaining = remainingQuotas[model] || 0;
                const percentage = totalQuota > 0 ? (used / totalQuota) * 100 : 0;
                
                return (
                  <div key={model} style={{ marginBottom: 12 }}>
                    <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                      <Text strong>{model}</Text>
                      <Text size="small">{used}/{totalQuota} (剩余{remaining})</Text>
                    </Space>
                    <Progress
                      percent={percentage}
                      size="small"
                      stroke={percentage > 80 ? '#f5222d' : percentage > 60 ? '#fa8c16' : '#52c41a'}
                      style={{ marginTop: 4 }}
                    />
                  </div>
                );
              })}
            </Space>
          </Card>
        </Space>
      );
    } catch (e) {
      return <Text type="danger">{t('数据解析错误')}</Text>;
    }
  };

  return (
    <>
      <Card
        title={
          <Space>
            <Users size={20} />
            <Text size="large" strong>
              {t('用户订阅管理')}
            </Text>
          </Space>
        }
        headerExtraContent={
          <Button
            theme="light"
            type="primary"
            icon={<IconRefresh />}
            onClick={() => loadSubscriptions()}
          >
            {t('刷新')}
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={subscriptions}
          pagination={{
            currentPage: activePage,
            pageSize: ITEMS_PER_PAGE,
            total: total,
            showSizeChanger: true,
            onPageChange: handlePageChange,
          }}
          loading={loading}
          empty={
            <Empty
              image={<IllustrationNoResult />}
              darkModeImage={<IllustrationNoResultDark />}
              description={t('暂无数据')}
            />
          }
        />
      </Card>

      <Modal
        title={t('订阅详情')}
        visible={showDetail}
        onCancel={() => setShowDetail(false)}
        footer={
          <Button onClick={() => setShowDetail(false)}>
            {t('关闭')}
          </Button>
        }
        width={800}
      >
        {renderSubscriptionDetail()}
      </Modal>
    </>
  );
};

export default UserSubscriptionsTable;
