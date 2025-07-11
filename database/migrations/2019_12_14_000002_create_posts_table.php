<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('posts', function (Blueprint $table) {
            $table->id();
            $table->string('title');
            $table->string('slug')->unique();
            $table->text('description')->nullable();
            $table->unsignedBigInteger('anilist_id')->nullable();
            $table->boolean('status')->default(false);
            $table->string('thumbnail')->nullable();
            $table->string('thumbnail_src')->nullable();
            $table->string('banner')->nullable();
            $table->string('banner_src')->nullable();
            $table->foreignId('year_id')->nullable()->constrained('years')->onDelete('set null');
            $table->foreignId('season_id')->nullable()->constrained('seasons')->onDelete('set null');
            $table->foreignId('format_id')->nullable()->constrained('formats')->onDelete('set null');
            $table->timestamps();
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('posts');
    }
};
