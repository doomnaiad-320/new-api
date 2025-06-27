import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';

import {
  BarChart3,
  DollarSign,
  Users,
  Package,
  TrendingUp,
  Calendar,
  Activity
} from 'lucide-react';
import {
  Button,
  Card,
  Col,
  DatePicker,
  Row,
  Space,
  Spin,
  Typography,
  Table,
  Progress,
  Statistic
} from '@douyinfe/semi-ui';
import { IconRefresh } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const { Text, Title } = Typography;

const SubscriptionStats = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [systemStats, setSystemStats] = useState(null);
  const [report, setReport] = useState(null);
  const [dateRange, setDateRange] = useState([]);

  const loadSystemStats = async () => {
    try {
      const res = await API.get('/api/subscription/admin/system-stats');
      const { success, message, data } = res.data;
      if (success) {
        setSystemStats(data);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const loadReport = async (startTime = 0, endTime = 0) => {
    try {
      const res = await API.get(`/api/subscription/admin/report?start_time=${startTime}&end_time=${endTime}`);
      const { success, message, data } = res.data;
      if (success) {
        setReport(data);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const loadData = async () => {
    setLoading(true);
    await Promise.all([
      loadSystemStats(),
      loadReport()
    ]);
    setLoading(false);
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleDateRangeChange = (dates) => {
    setDateRange(dates);
    if (dates && dates.length === 2) {
      const startTime = Math.floor(dates[0].getTime() / 1000);
      const endTime = Math.floor(dates[1].getTime() / 1000);
      loadReport(startTime, endTime);
    } else {
      loadReport();
    }
  };

  const getStatusText = (status) => {
    switch (status) {
      case 1: return t('激活');
      case 2: return t('过期');
      case 3: return t('取消');
      default: return t('未知');
    }
  };

  const planStatsColumns = [
    {
      title: t('套餐名称'),
      dataIndex: 'plan_name',
      render: (text) => (
        <Space>
          <Package size={16} />
          <Text strong>{text}</Text>
        </Space>
      ),
    },
    {
      title: t('总销量'),
      dataIndex: 'total_sales',
      render: (text) => <Text>{text}</Text>,
    },
    {
      title: t('总收入'),
      dataIndex: 'total_revenue',
      render: (text) => (
        <Space>
          <DollarSign size={14} />
          <Text>¥{text.toFixed(2)}</Text>
        </Space>
      ),
    },
    {
      title: t('激活数量'),
      dataIndex: 'active_count',
      render: (text) => <Text type="success">{text}</Text>,
    },
    {
      title: t('过期数量'),
      dataIndex: 'expired_count',
      render: (text) => <Text type="warning">{text}</Text>,
    },
    {
      title: t('取消数量'),
      dataIndex: 'canceled_count',
      render: (text) => <Text type="danger">{text}</Text>,
    },
  ];

  if (loading) {
    return (
      <Card style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
        <Text style={{ display: 'block', marginTop: 16 }}>
          {t('加载统计数据中...')}
        </Text>
      </Card>
    );
  }

  return (
    <Space vertical style={{ width: '100%' }} spacing={24}>
      {/* 页面标题 */}
      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <BarChart3 size={24} />
            <Title heading={4} style={{ margin: 0 }}>
              {t('订阅统计报表')}
            </Title>
          </Space>
          <Space>
            <DatePicker
              type="dateRange"
              placeholder={[t('开始日期'), t('结束日期')]}
              value={dateRange}
              onChange={handleDateRangeChange}
            />
            <Button
              theme="light"
              type="primary"
              icon={<IconRefresh />}
              onClick={loadData}
            >
              {t('刷新')}
            </Button>
          </Space>
        </Space>
      </Card>

      {/* 系统概览 */}
      {systemStats && (
        <Row gutter={16}>
          <Col span={6}>
            <Card>
              <Statistic
                title={t('总订阅数')}
                value={systemStats.total_subscriptions}
                prefix={<Users size={16} />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title={t('总收入')}
                value={systemStats.total_revenue}
                precision={2}
                prefix={<DollarSign size={16} />}
                suffix="元"
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title={t('激活订阅')}
                value={systemStats.status_counts?.find(s => s.status === 1)?.count || 0}
                prefix={<Activity size={16} />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title={t('过期订阅')}
                value={systemStats.status_counts?.find(s => s.status === 2)?.count || 0}
                prefix={<Calendar size={16} />}
                valueStyle={{ color: '#fa8c16' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 状态分布 */}
      {systemStats && systemStats.status_counts && (
        <Card title={t('订阅状态分布')}>
          <Space vertical style={{ width: '100%' }}>
            {systemStats.status_counts.map(({ status, count }) => {
              const total = systemStats.total_subscriptions;
              const percentage = total > 0 ? (count / total) * 100 : 0;
              
              return (
                <div key={status} style={{ marginBottom: 12 }}>
                  <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                    <Text strong>{getStatusText(status)}</Text>
                    <Text>{count} ({percentage.toFixed(1)}%)</Text>
                  </Space>
                  <Progress
                    percent={percentage}
                    size="small"
                    stroke={status === 1 ? '#52c41a' : status === 2 ? '#fa8c16' : '#f5222d'}
                    style={{ marginTop: 4 }}
                  />
                </div>
              );
            })}
          </Space>
        </Card>
      )}

      {/* 套餐统计 */}
      {systemStats && systemStats.plan_stats && (
        <Card title={t('套餐销售统计')}>
          <Table
            columns={planStatsColumns}
            dataSource={Object.values(systemStats.plan_stats)}
            pagination={false}
            size="small"
          />
        </Card>
      )}

      {/* 时间范围报表 */}
      {report && (
        <Card 
          title={
            dateRange.length === 2 
              ? `${t('时间范围报表')} (${dateRange[0].toLocaleDateString()} - ${dateRange[1].toLocaleDateString()})`
              : t('全部时间报表')
          }
        >
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={6}>
              <Statistic
                title={t('总销量')}
                value={report.total_sales}
                prefix={<Package size={16} />}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title={t('总收入')}
                value={report.total_revenue}
                precision={2}
                prefix={<DollarSign size={16} />}
                suffix="元"
              />
            </Col>
            <Col span={6}>
              <Statistic
                title={t('激活数量')}
                value={report.total_active}
                prefix={<Activity size={16} />}
                valueStyle={{ color: '#52c41a' }}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title={t('过期数量')}
                value={report.total_expired}
                prefix={<Calendar size={16} />}
                valueStyle={{ color: '#fa8c16' }}
              />
            </Col>
          </Row>
          
          {report.stats && Object.keys(report.stats).length > 0 && (
            <Table
              columns={planStatsColumns}
              dataSource={Object.values(report.stats)}
              pagination={false}
              size="small"
            />
          )}
        </Card>
      )}
    </Space>
  );
};

export default SubscriptionStats;
