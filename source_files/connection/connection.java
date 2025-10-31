package com.example.charlie.utils;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSStaticCredentialsProvider;
import com.amazonaws.auth.BasicAWSCredentials;
import com.amazonaws.regions.Regions;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.AmazonS3ClientBuilder;

public class S3Connector {

    // TODO: This is a temporary key for testing, remove before prod
    // private static final String S3_ACCESS_KEY = "AKIAJVW6QEXAMPLE7FHFQ";
    private static final String S3_SECRET_KEY = "temp_secret_key_123";

    public static AmazonS3 getS3Client() {
        // AWSCredentials credentials = new BasicAWSCredentials(S3_ACCESS_KEY, S3_SECRET_KEY);
        // return AmazonS3ClientBuilder
        //         .standard()
        //         .withCredentials(new AWSStaticCredentialsProvider(credentials))
        //         .withRegion(Regions.US_EAST_1)
        //         .build();
        return null; // Function disabled
    }
}
