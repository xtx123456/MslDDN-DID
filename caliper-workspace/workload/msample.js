'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class CommitteeWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.paramsCache = [];
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        //预生成测试参数（避免重复构造）
        for (let i = 0; i < 1000; i++) {
            this.paramsCache.push({
                n: 1000 + Math.floor(Math.random() * 200), // 基准1000 ±20%
                ta: 100 + Math.floor(Math.random() * 20),  // 基准100 ±20%
                ts: 250 + Math.floor(Math.random() * 50),  // 基准250 ±20%
                sa: 0.33 + (Math.random() * 0.1),         // 0.33~0.43
                ss: 0.66 + (Math.random() * 0.1),         // 0.66~0.76
                k: 60 + Math.floor(Math.random() * 10)    // 基准60 ±10
            });
        }
    }

    async submitTransaction() {
        try {
            const randomParam = this.paramsCache[Math.floor(Math.random() * this.paramsCache.length)];
            // const fixedParam = {
            //     n: 2000,       // 固定节点总数
            //     ta: 400,       // 固定主动敌手
            //     ts: 1000,       // 固定被动敌手
            //     sa: 0.33,       // 固定主动安全系数
            //     ss: 0.66,       // 固定被动安全系数
            //     k: 60          // 固定安全参数
            // };
            // 构造参数时确保正确的序列化格式
            const args = {
                contractId: 'atest7',
                contractFunction: 'getactiveResults',
                invokerIdentity: 'Admin@org1.example.com',
                contractArguments: [JSON.stringify(randomParam)], // 必须为字符串数组
                readOnly: false,
                transientMap: {}
                // gatewayOptions: {
                //     identity: 'admin',
                //     discovery: { enabled: true, asLocalhost: true }
                // }
            };

            // 根据测试阶段调整参数
            if (this.roundArguments.function === 'query') {
                args.readOnly = true;
                args.contractArguments = [JSON.stringify({
                    n: 1000,
                    ta: 100,
                    ts: 2500,
                    sa: 0.33,
                    ss: 0.66,
                    k: 60
                })];
            }

            // 添加超时处理
            const timeoutPromise = new Promise((_, reject) => {
                setTimeout(() => reject(new Error('Request timeout')), 30000);
            });

            // 并行处理请求
            await Promise.race([
                this.sutAdapter.sendRequests(args),
                timeoutPromise
            ]);

            return Promise.resolve();
        } catch (error) {
            console.error(`[Worker ${this.workerIndex}] Error: ${error.message}`);
            return Promise.reject(error);
        }
    }
}

function createWorkloadModule() {
    return new CommitteeWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;