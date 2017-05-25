# eventbus
[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex023/eventbus)](https://goreportcard.com/report/github.com/alex023/eventbus)
[![GoDoc](https://godoc.org/github.com/alex023/eventbus?status.svg)](https://godoc.org/github.com/alex023/eventbus)
[![Build Status](https://travis-ci.org/alex023/eventbus.svg?branch=dev)](https://travis-ci.org/alex023/eventbus?branch=dev)
[![Coverage Status](https://coveralls.io/repos/github/alex023/eventbus/badge.svg?branch=dev)](https://coveralls.io/github/alex023/eventbus?branch=dev)


Pub-sub systems based on asynchronous models provide filter support for each topic to meet flexible business support.

eventbus 是一个加强版的单机pub-sub异步消息系统。系统提供大吞吐量的单机消息引擎框架。除了一般的消息订阅、发布的功能支持之外，
 也适用于简单的ECS应用场景。具有以下特征：
- 大吞吐量，不低于200wQPS
- 隔离业务崩溃
- 主题[过滤器]支持,且可以在任意时间添加或卸载
- [todo]考虑提供统计插件支持