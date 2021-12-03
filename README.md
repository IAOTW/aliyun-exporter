# Aliyun CloudMonitor Exporter

exporter for Aliyun CloudMonitor. Written in Golang.
inspired by [aliyun-exporter](https://github.com/fengxsong/aliyun-exporter)

## Develop

```bash
cd aliyun-exporter
make tidy
```

## Usage

```bash
# build
go build -o aliyun-exporter
# run
./aliyun-exporter serve
```

### prometheus job

```yaml
- job_name: 'aliyun-exporter'
  scrape_interval: 60s
  scrape_timeout: 60s
  static_configs:
  - targets: ['aliyun-exporter:9527']
    labels:
      account_name: xxxx
      provider: aliyun # or aliyun_jst
```

### dashboard
[Instance Dashboard](https://../dashboards/Aliyun-Instance-Dashboard.json)

## Limitation

- an exporter instance can only scrape metrics from single region
- ...

## TODO

- dynamic rate limiter

## Ref

- https://partners-intl.aliyun.com/help/doc-detail/51939.htm?spm=a2c63.p38356.b99.150.7c8312d7lwqVVC
- https://partners-intl.aliyun.com/help/doc-detail/163515.htm?spm=a2c63.p38356.a3.4.326357304ihN0P
- https://github.com/aliyun/alibaba-cloud-sdk-go