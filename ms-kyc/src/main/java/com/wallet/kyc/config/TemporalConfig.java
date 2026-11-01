package com.wallet.kyc.config;

import com.wallet.kyc.temporal.activity.*;
import com.wallet.kyc.temporal.workflow.KycOnboardingWorkflowImpl;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowClientOptions;
import io.temporal.serviceclient.WorkflowServiceStubs;
import io.temporal.serviceclient.WorkflowServiceStubsOptions;
import io.temporal.worker.Worker;
import io.temporal.worker.WorkerFactory;
import jakarta.annotation.PreDestroy;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class TemporalConfig {

    private static final Logger log = LoggerFactory.getLogger(TemporalConfig.class);

    @Value("${temporal.target:localhost:7233}")
    private String temporalTarget;

    @Value("${temporal.namespace:default}")
    private String namespace;

    @Value("${temporal.task-queue:KYC_ONBOARDING_TASK_QUEUE}")
    private String taskQueue;

    private WorkerFactory workerFactory;

    @Bean
    public WorkflowServiceStubs workflowServiceStubs() {
        log.info("Connecting to Temporal server at {}", temporalTarget);
        WorkflowServiceStubsOptions options = WorkflowServiceStubsOptions.newBuilder()
                .setTarget(temporalTarget)
                .build();
        return WorkflowServiceStubs.newServiceStubs(options);
    }

    @Bean
    public WorkflowClient workflowClient(WorkflowServiceStubs workflowServiceStubs) {
        WorkflowClientOptions options = WorkflowClientOptions.newBuilder()
                .setNamespace(namespace)
                .build();
        return WorkflowClient.newInstance(workflowServiceStubs, options);
    }

    @Bean
    @ConditionalOnProperty(name = "temporal.worker.enabled", havingValue = "true", matchIfMissing = true)
    public WorkerFactory workerFactory(WorkflowClient workflowClient,
                                       DocumentVerificationActivity documentVerificationActivity,
                                       BiometricVerificationActivity biometricVerificationActivity,
                                       SanctionsScreeningActivity sanctionsScreeningActivity,
                                       RiskAssessmentActivity riskAssessmentActivity,
                                       KycPersistenceActivity kycPersistenceActivity,
                                       NotificationActivity notificationActivity) {
        log.info("Registering Temporal Worker for task queue: {}", taskQueue);
        this.workerFactory = WorkerFactory.newInstance(workflowClient);
        Worker worker = workerFactory.newWorker(taskQueue);

        worker.registerWorkflowImplementationTypes(KycOnboardingWorkflowImpl.class);
        worker.registerActivitiesImplementations(
                documentVerificationActivity,
                biometricVerificationActivity,
                sanctionsScreeningActivity,
                riskAssessmentActivity,
                kycPersistenceActivity,
                notificationActivity
        );

        workerFactory.start();
        log.info("Temporal WorkerFactory started successfully for task queue: {}", taskQueue);
        return workerFactory;
    }

    @PreDestroy
    public void shutdown() {
        if (workerFactory != null) {
            log.info("Shutting down Temporal WorkerFactory...");
            workerFactory.shutdown();
        }
    }
}
